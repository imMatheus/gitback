package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type AnalyzeRequest struct {
	Username string `json:"username"`
	Repo     string `json:"repo"`
}

type AnalyzeResponse struct {
	Message           string `json:"message"`
	TotalAdded        int    `json:"totalAdded"`
	TotalRemoved      int    `json:"totalRemoved"`
	TotalContributors int    `json:"totalContributors"`
}

func main() {
	// Get API URL from environment or use default
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	// List of popular GitHub repositories
	repos := []struct {
		Username string
		Repo     string
	}{
		// {"immatheus", "css-subway-surfer"},
		// {"microsoft", "vscode"},
		// {"facebook", "react"},
		// {"microsoft", "TypeScript"},
		// {"vercel", "next.js"},
		// {"kubernetes", "kubernetes"},
		// {"tensorflow", "tensorflow"},
		// {"golang", "go"},
		// {"rust-lang", "rust"},
		// {"nodejs", "node"},
		// {"pytorch", "pytorch"},
		// {"django", "django"},
		// {"ansible", "ansible"},
		// {"elastic", "elasticsearch"},
		// {"apache", "spark"},
		// {"rails", "rails"},
		// {"vuejs", "vue"},
		// {"angular", "angular"},
		// {"dotnet", "aspnetcore"},
		// {"docker", "compose"},
		// {"airbnb", "javascript"},
		// {"mui", "material-ui"},
		// {"axios", "axios"},
		// {"mrdoob", "three.js"},
		// {"lodash", "lodash"},
		// {"moment", "moment"},
		// {"chartjs", "Chart.js"},
		// {"spring-projects", "spring-boot"},
		// {"laravel", "laravel"},
		// {"expressjs", "express"},
		// {"nestjs", "nest"},
		// {"sveltejs", "svelte"},
		// {"nuxt", "nuxt"},
		// {"remix-run", "remix"},
		// {"fastapi", "fastapi"},
		// {"gin-gonic", "gin"},
		// {"gofiber", "fiber"},
		// {"torvalds", "linux"},
	}

	fmt.Printf("Filling cache with %d popular repositories...\n", len(repos))
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Running 2 requests in parallel\n\n")

	client := &http.Client{
		Timeout: 10 * time.Minute, // Some repos might take a while
	}

	var (
		successCount int
		failCount    int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	// Semaphore to limit concurrency to 5
	sem := make(chan struct{}, 8)

	for i, repo := range repos {
		wg.Add(1)
		go func(index int, r struct {
			Username string
			Repo     string
		}) {
			defer wg.Done()

			// Acquire semaphore (blocks if 2 are already running)
			sem <- struct{}{}
			defer func() { <-sem }() // Release semaphore when done

			fmt.Printf("[%d/%d] Analyzing %s/%s...\n", index+1, len(repos), r.Username, r.Repo)

			reqBody := AnalyzeRequest{
				Username: r.Username,
				Repo:     r.Repo,
			}

			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				log.Printf("Error marshaling request: %v", err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/analyze", apiURL), bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error creating request: %v", err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			req.Header.Set("Content-Type", "application/json")

			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error making request: %v", err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			duration := time.Since(start)

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				var result AnalyzeResponse
				if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
					fmt.Printf("  ✓ Success! (%d lines, %d contributors) - %v\n",
						result.TotalAdded-result.TotalRemoved, result.TotalContributors, duration)
					successCount++
				} else {
					fmt.Printf("  ✓ Success! (cached or processing) - %v\n", duration)
					successCount++
				}
			} else if resp.StatusCode == http.StatusNotFound {
				fmt.Printf("  ✗ Repository not found (404)\n")
				failCount++
			} else {
				fmt.Printf("  ✗ Failed with status %d - %v\n", resp.StatusCode, duration)
				failCount++
			}
			mu.Unlock()
		}(i, repo)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Failed:  %d\n", failCount)
	fmt.Printf("Total:   %d\n", len(repos))
}
