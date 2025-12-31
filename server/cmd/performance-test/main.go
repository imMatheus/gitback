package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type AnalyzeRequest struct {
	Username string `json:"username"`
	Repo     string `json:"repo"`
}

type TestResult struct {
	Duration     time.Duration
	StatusCode   int
	Success      bool
	Error        string
	ResponseSize int64
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <server-url> [concurrent-requests] [test-repo]")

	}

	serverURL := os.Args[1]
	concurrentRequests := 10
	testRepo := "facebook/react" // Default test repository

	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &concurrentRequests)
	}
	if len(os.Args) > 3 {
		testRepo = os.Args[3]
	}

	log.Printf("Starting performance test against %s", serverURL)
	log.Printf("Concurrent requests: %d", concurrentRequests)
	log.Printf("Test repository: %s", testRepo)

	// Parse repository name
	parts := strings.Split(testRepo, "/")

	request := AnalyzeRequest{
		Username: parts[0],
		Repo:     parts[1],
	}

	// Warm up the server
	log.Println("Warming up server...")
	warmupResult := performRequest(serverURL, request)
	log.Printf("Warmup completed: %v (Status: %d)", warmupResult.Duration, warmupResult.StatusCode)

	// Wait a moment for cache to settle
	time.Sleep(2 * time.Second)

	// Run concurrent performance test
	log.Printf("Running %d concurrent requests...", concurrentRequests)
	results := runConcurrentTest(serverURL, request, concurrentRequests)

	// Analyze results
	analyzeResults(results)
}

func performRequest(serverURL string, req AnalyzeRequest) TestResult {
	start := time.Now()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return TestResult{
			Duration: time.Since(start),
			Success:  false,
			Error:    fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	resp, err := http.Post(serverURL+"/api/analyze", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return TestResult{
			Duration: time.Since(start),
			Success:  false,
			Error:    fmt.Sprintf("HTTP request error: %v", err),
		}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	// Read response to get accurate timing
	buf := new(bytes.Buffer)
	responseSize, _ := buf.ReadFrom(resp.Body)

	return TestResult{
		Duration:     duration,
		StatusCode:   resp.StatusCode,
		Success:      resp.StatusCode == 200,
		ResponseSize: responseSize,
		Error:        "",
	}
}

func runConcurrentTest(serverURL string, req AnalyzeRequest, concurrency int) []TestResult {
	var wg sync.WaitGroup
	results := make([]TestResult, concurrency)

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = performRequest(serverURL, req)
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	log.Printf("All %d requests completed in %v", concurrency, totalTime)
	return results
}

func analyzeResults(results []TestResult) {
	var (
		totalDuration time.Duration
		successCount  int
		errorCount    int
		minDuration   = time.Hour
		maxDuration   time.Duration
		totalSize     int64
		statusCodes   = make(map[int]int)
	)

	log.Println("\n=== PERFORMANCE TEST RESULTS ===")

	for i, result := range results {
		log.Printf("Request %d: %v (Status: %d, Size: %d bytes, Success: %t)",
			i+1, result.Duration, result.StatusCode, result.ResponseSize, result.Success)

		totalDuration += result.Duration
		totalSize += result.ResponseSize
		statusCodes[result.StatusCode]++

		if result.Success {
			successCount++
		} else {
			errorCount++
			if result.Error != "" {
				log.Printf("  Error: %s", result.Error)
			}
		}

		if result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	avgDuration := totalDuration / time.Duration(len(results))
	avgSize := totalSize / int64(len(results))

	log.Println("\n=== SUMMARY ===")
	log.Printf("Total Requests: %d", len(results))
	log.Printf("Successful Requests: %d (%.1f%%)", successCount, float64(successCount)/float64(len(results))*100)
	log.Printf("Failed Requests: %d (%.1f%%)", errorCount, float64(errorCount)/float64(len(results))*100)
	log.Printf("Average Response Time: %v", avgDuration)
	log.Printf("Min Response Time: %v", minDuration)
	log.Printf("Max Response Time: %v", maxDuration)
	log.Printf("Average Response Size: %d bytes (%.2f KB)", avgSize, float64(avgSize)/1024)
	log.Printf("Total Data Transferred: %d bytes (%.2f MB)", totalSize, float64(totalSize)/1024/1024)

	log.Println("\n=== STATUS CODES ===")
	for code, count := range statusCodes {
		log.Printf("HTTP %d: %d requests (%.1f%%)", code, count, float64(count)/float64(len(results))*100)
	}

	// Performance Assessment
	log.Println("\n=== PERFORMANCE ASSESSMENT ===")
	if avgDuration < 200*time.Millisecond {
		log.Println("✅ Excellent performance: Sub-200ms average response time")
	} else if avgDuration < 1*time.Second {
		log.Println("✅ Good performance: Sub-1s average response time")
	} else if avgDuration < 5*time.Second {
		log.Println("⚠️  Acceptable performance: Sub-5s average response time")
	} else {
		log.Println("❌ Poor performance: >5s average response time - optimization needed")
	}

	if successCount == len(results) {
		log.Println("✅ Perfect reliability: 100% success rate")
	} else if float64(successCount)/float64(len(results)) > 0.95 {
		log.Println("✅ Good reliability: >95% success rate")
	} else {
		log.Println("❌ Poor reliability: <95% success rate - stability issues detected")
	}
}
