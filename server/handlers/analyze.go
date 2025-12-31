package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	database "github.com/immatheus/gitback/databases"
	"github.com/immatheus/gitback/git"
	"github.com/immatheus/gitback/middleware"
	"github.com/immatheus/gitback/storage"
)

type AnalyzeRequest struct {
	Username string `json:"username" validate:"required,min=1,max=255"`
	Repo     string `json:"repo" validate:"required,min=1,max=255"`
}

type GitHubRepo struct {
	StargazersCount int    `json:"stargazers_count"`
	Language        string `json:"language"`
	Size            int    `json:"size"`
}

type GitHubPullRequest struct {
	ID          int64                  `json:"id"`
	Number      int                    `json:"number"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	User        GitHubUser             `json:"user"`
	CreatedAt   string                 `json:"created_at"`
	State       string                 `json:"state"`
	HTMLURL     string                 `json:"html_url"`
	Comments    int                    `json:"comments"`
	PullRequest *GitHubPullRequestInfo `json:"pull_request,omitempty"`
	Reactions   GitHubReactions        `json:"reactions"`
}

type GitHubPullRequestInfo struct {
	MergedAt *string `json:"merged_at"`
}

type GitHubReactions struct {
	TotalCount int `json:"total_count"`
	PlusOne    int `json:"+1"`
	MinusOne   int `json:"-1"`
	Laugh      int `json:"laugh"`
	Hooray     int `json:"hooray"`
	Confused   int `json:"confused"`
	Heart      int `json:"heart"`
	Rocket     int `json:"rocket"`
	Eyes       int `json:"eyes"`
}

type GitHubUser struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type GitHubSearchResult struct {
	TotalCount int                 `json:"total_count"`
	Items      []GitHubPullRequest `json:"items"`
}

var githubClient = &http.Client{
	Timeout: 15 * time.Second,
}

func AnalyzeRepo(c *fiber.Ctx) error {
	requestStart := time.Now()

	var req AnalyzeRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return middleware.ValidationError(c, "Invalid request body")
	}

	// Validate input
	if err := validateRequest(req); err != nil {
		return middleware.ValidationError(c, err.Error())
	}

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", req.Username, req.Repo)
	log.Printf("=== Starting analysis for: %s ===", repoURL)

	if cachedData, err := storage.GetFromCache(req.Username, req.Repo); err != nil {
		log.Printf("Cache check failed: %v", err)
	} else if cachedData != nil {
		log.Printf("Returning cached analysis for %s", repoURL)

		// Update view count in background
		go func() {
			if err := database.IncrementViews(req.Username, req.Repo); err != nil {
				log.Printf("[DB] Failed to increment views for %s: %v", repoURL, err)
			}
		}()

		return c.JSON(cachedData)
	}

	// Validate repository URL before processing
	if err := git.ValidateRepoURL(repoURL); err != nil {
		return middleware.ValidationError(c, err.Error())
	}

	// Clone and analyze repository with improved git operations
	repo, err := git.CloneRepository(repoURL)
	if err != nil {
		if isNotFoundError(err) {
			log.Printf("Repository not found: %s - Error: %v", repoURL, err)
			return middleware.NotFoundError(c, "Repository not found")
		}
		log.Printf("Failed to clone repository: %s - Error: %v", repoURL, err)
		return middleware.InternalError(c, "Failed to clone repository")
	}
	defer repo.Cleanup()

	commits, err := repo.AnalyzeCommits()
	if err != nil {
		log.Printf("Failed to analyze commits for %s: %v", repoURL, err)
		return middleware.InternalError(c, "Failed to analyze repository")
	}

	// Process statistics
	totalAdded := 0
	totalRemoved := 0
	totalContributors := 0
	contributors := make(map[string]bool)

	for _, commit := range commits {
		totalAdded += commit.Added
		totalRemoved += commit.Removed
		if _, ok := contributors[commit.Author]; !ok {
			contributors[commit.Author] = true
			totalContributors++
		}
	}

	log.Printf("Analysis completed for %s: %d commits, %d contributors, +%d/-%d lines",
		repoURL, len(commits), totalContributors, totalAdded, totalRemoved)

	// Fetch GitHub data in parallel
	var githubInfo *GitHubRepo
	var pullRequests *GitHubSearchResult

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if repoInfo, err := fetchGitHubRepoInfo(req.Username, req.Repo); err == nil {
			githubInfo = repoInfo
		} else {
			log.Printf("Failed to fetch GitHub repo info: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if pullRequestInfo, err := fetchRepoTopPullRequests(req.Username, req.Repo); err == nil {
			pullRequests = pullRequestInfo
		} else {
			log.Printf("Failed to fetch top pull requests: %v", err)
		}
	}()

	wg.Wait()

	// Save to database in background
	go func() {
		histogram := database.CalculateLinesHistogram(commits, 10)
		totalLines := totalAdded - totalRemoved

		dbData := database.RepoData{
			Username:       req.Username,
			RepoName:       req.Repo,
			TotalAdditions: totalAdded,
			TotalLines:     totalLines,
			TotalRemovals:  totalRemoved,
			LinesHistogram: histogram,
			TotalStars:     githubInfo.StargazersCount,
			TotalCommits:   len(commits),
			Language:       githubInfo.Language,
			Size:           githubInfo.Size,
		}

		if err := database.SaveRepo(dbData); err != nil {
			log.Printf("[DB] Failed to save repo to database for %s: %v", repoURL, err)
		}

		if err := database.IncrementViews(req.Username, req.Repo); err != nil {
			log.Printf("[DB] Failed to increment views for %s: %v", repoURL, err)
		}
	}()

	response := fiber.Map{
		"totalAdded":        totalAdded,
		"totalRemoved":      totalRemoved,
		"totalContributors": totalContributors,
		"totalCommits":      len(commits),
		"commits":           commits,
		"github":            githubInfo,
		"pullRequests":      pullRequests,
	}

	// Store in cache asynchronously
	go func() {
		if err := storage.StoreInCache(req.Username, req.Repo, response); err != nil {
			log.Printf("Failed to store analysis in cache for %s: %v", repoURL, err)
		}
	}()

	log.Printf("[TIMING] Total request time: %v", time.Since(requestStart))
	return c.JSON(response)
}

func validateRequest(req AnalyzeRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Repo == "" {
		return fmt.Errorf("repo is required")
	}

	// Validate against potential injection
	if containsUnsafeChars(req.Username) || containsUnsafeChars(req.Repo) {
		return fmt.Errorf("invalid characters in repository name")
	}

	return nil
}

func containsUnsafeChars(s string) bool {
	return strings.ContainsAny(s, ";|&$`(){}[]<>\"'")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "exit status 128") ||
		strings.Contains(errStr, "Repository not found") ||
		strings.Contains(errStr, "fatal: repository") ||
		strings.Contains(errStr, "remote: Repository not found")
}

func fetchGitHubRepoInfo(username, repo string) (*GitHubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", username, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := githubClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var repoInfo GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}

	return &repoInfo, nil
}

func fetchRepoTopPullRequests(username, repo string) (*GitHubSearchResult, error) {
	const prCount = 5
	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=repo:%s/%s+type:pr+created:2025-01-01..2025-12-31&sort=reactions&order=desc&per_page=%d", username, repo, prCount)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := githubClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var searchResult GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	return &searchResult, nil
}
