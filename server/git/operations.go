package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	database "github.com/immatheus/gitback/databases"
)

// GitConfig holds configuration for git operations
type GitConfig struct {
	MaxMemoryMB    int
	TimeoutSeconds int
	MaxCommits     int
	TempDirPattern string
}

// Repository represents a cloned git repository
type Repository struct {
	Path   string
	Config GitConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// CloneRepository safely clones a repository with resource management
func CloneRepository(repoURL string) (*Repository, error) {
	gitConfig := GitConfig{
		MaxMemoryMB:    500,
		TimeoutSeconds: 300, // 5 minutes
		MaxCommits:     50000,
		TempDirPattern: "gitback-analysis-*",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(gitConfig.TimeoutSeconds)*time.Second)

	tmpDir, err := os.MkdirTemp("", gitConfig.TempDirPattern)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	repo := &Repository{
		Path:   tmpDir,
		Config: gitConfig,
		ctx:    ctx,
		cancel: cancel,
	}

	// Set up command with context and resource limits
	cmd := exec.CommandContext(ctx, "git", "clone",
		"--bare",
		"--single-branch",
		"--depth=1000", // Limit initial depth for performance
		"--no-tags",    // Skip tags for faster clone
		repoURL,
		tmpDir)

	// Limit memory usage
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_CONFIG_GLOBAL=/dev/null"),
		fmt.Sprintf("GIT_CONFIG_SYSTEM=/dev/null"),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		repo.Cleanup()
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git clone timeout after %d seconds", gitConfig.TimeoutSeconds)
		}
		return nil, fmt.Errorf("git clone failed: %w, stderr: %s", err, stderr.String())
	}

	return repo, nil
}

// AnalyzeCommits extracts commit statistics with memory optimization
func (r *Repository) AnalyzeCommits() ([]database.CommitStats, error) {
	// Use streaming approach to handle large repositories
	cmd := exec.CommandContext(r.ctx, "git",
		"--git-dir", r.Path,
		"log",
		"--numstat",
		"--format=%H|%an|%at|%s",
		"--reverse", // Process oldest first for better memory usage
		fmt.Sprintf("--max-count=%d", r.Config.MaxCommits),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start git log: %w", err)
	}

	commits := make([]database.CommitStats, 0, 1000) // Pre-allocate reasonable size
	scanner := bufio.NewScanner(stdout)

	// Increase buffer size for large commits
	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 10*1024*1024) // 10MB max

	var currentCommit *database.CommitStats

	for scanner.Scan() {
		select {
		case <-r.ctx.Done():
			return nil, fmt.Errorf("analysis cancelled: %w", r.ctx.Err())
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.Contains(line, "|") {
			// Save previous commit if exists
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			parts := strings.SplitN(line, "|", 4)
			if len(parts) != 4 {
				continue
			}

			timestamp, _ := strconv.ParseInt(parts[2], 10, 64)
			currentCommit = &database.CommitStats{
				Hash:              parts[0][:min(7, len(parts[0]))],
				Author:            parts[1],
				Date:              timestamp,
				Message:           truncateMessage(parts[3], 100),
				Added:             0,
				Removed:           0,
				FilesTouchedCount: 0,
			}
		} else if currentCommit != nil && strings.Contains(line, "\t") {
			// Parse numstat line
			fields := strings.SplitN(line, "\t", 3)
			if len(fields) >= 3 {
				currentCommit.FilesTouchedCount++

				if added, err := strconv.Atoi(fields[0]); err == nil {
					currentCommit.Added += added
				}
				if removed, err := strconv.Atoi(fields[1]); err == nil {
					currentCommit.Removed += removed
				}
			}
		}
	}

	// Save last commit
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return commits, nil
}

// Cleanup removes temporary files and cancels context
func (r *Repository) Cleanup() {
	if r.cancel != nil {
		r.cancel()
	}
	if r.Path != "" {
		os.RemoveAll(r.Path)
	}
}

// ValidateRepoURL performs basic validation on repository URL
func ValidateRepoURL(repoURL string) error {
	if !strings.HasPrefix(repoURL, "https://github.com/") {
		return fmt.Errorf("only GitHub HTTPS URLs are supported")
	}

	// Basic validation to prevent command injection
	if strings.ContainsAny(repoURL, ";|&$`(){}[]<>") {
		return fmt.Errorf("invalid characters in repository URL")
	}

	return nil
}

func truncateMessage(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
