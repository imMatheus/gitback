package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "GitBack v1.0.0",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "GitBack API",
			"version": "1.0.0",
		})
	})

	app.Post("/api/analyze", analyzeRepo)

	// Get port from environment (Cloud Run sets this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log system info on startup
	logSystemInfo()

	log.Fatal(app.Listen(":" + port))
}

func logSystemInfo() {
	log.Printf("=== SYSTEM INFO ===")
	log.Printf("NumCPU: %d", runtime.NumCPU())
	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS(0))
	log.Printf("GOOS: %s", runtime.GOOS)
	log.Printf("GOARCH: %s", runtime.GOARCH)

	// Check git version
	gitVersion := exec.Command("git", "--version")
	out, err := gitVersion.Output()
	if err == nil {
		log.Printf("Git version: %s", strings.TrimSpace(string(out)))
	}

	// Check git config
	protocolCmd := exec.Command("git", "config", "--get", "protocol.version")
	protocolOut, _ := protocolCmd.Output()
	log.Printf("Git protocol.version: %s", strings.TrimSpace(string(protocolOut)))

	// Test network speed to github.com
	log.Printf("Testing network speed to github.com...")
	testNetworkSpeed()

	log.Printf("===================")
}

func testNetworkSpeed() {
	start := time.Now()
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{time_total}", "https://github.com")
	output, err := cmd.Output()
	elapsed := time.Since(start)

	if err == nil {
		log.Printf("GitHub ping: %s (total time: %v)", strings.TrimSpace(string(output)), elapsed)
	} else {
		log.Printf("GitHub ping failed: %v", err)
	}
}

func analyzeRepo(c *fiber.Ctx) error {
	requestStart := time.Now()

	type Request struct {
		Username string `json:"username"`
		Repo     string `json:"repo"`
	}

	parseStart := time.Now()
	var req Request
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	log.Printf("[TIMING] Body parsing: %v", time.Since(parseStart))

	if req.Username == "" {
		log.Printf("Missing username in request")
		return c.Status(400).JSON(fiber.Map{
			"error": "username is required",
		})
	}

	if req.Repo == "" {
		log.Printf("Missing repo in request")
		return c.Status(400).JSON(fiber.Map{
			"error": "repo is required",
		})
	}

	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", req.Username, req.Repo)
	log.Printf("=== Starting analysis for: %s ===", repoURL)

	// Log system stats at request time
	logRequestSystemStats()

	cloneStart := time.Now()
	commits, fileTouchCounts, err := cloneRepo(repoURL)
	cloneDuration := time.Since(cloneStart)
	log.Printf("[TIMING] Total cloneRepo execution: %v", cloneDuration)

	if err != nil {
		if isNotFoundError(err) {
			log.Printf("Repository not found: %s - Error: %v", repoURL, err)
			return c.Status(404).JSON(fiber.Map{
				"error": "Repository not found",
			})
		}
		log.Printf("Failed to clone repository: %s - Error: %v", repoURL, err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to clone repository",
		})
	}

	processingStart := time.Now()
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
	log.Printf("[TIMING] Processing commits stats: %v", time.Since(processingStart))

	log.Printf("Analysis completed for %s: %d commits, %d contributors, +%d/-%d lines",
		repoURL, len(commits), totalContributors, totalAdded, totalRemoved)

	topFilesStart := time.Now()
	topFiles := getTopTouchedFiles(fileTouchCounts, 100)
	log.Printf("[TIMING] Get top touched files: %v", time.Since(topFilesStart))

	jsonStart := time.Now()
	response := fiber.Map{
		"message":           "Analysis completed",
		"totalAdded":        totalAdded,
		"totalRemoved":      totalRemoved,
		"totalContributors": totalContributors,
		"commits":           commits,
		"mostTouchedFiles":  topFiles,
	}

	// Measure JSON marshaling time
	err = c.JSON(response)
	jsonDuration := time.Since(jsonStart)
	log.Printf("[TIMING] JSON marshaling and response: %v", jsonDuration)

	totalDuration := time.Since(requestStart)
	log.Printf("[TIMING] *** TOTAL REQUEST TIME: %v ***", totalDuration)
	log.Printf("=== Request completed ===\n")

	return err
}

func logRequestSystemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("[STATS] Goroutines: %d", runtime.NumGoroutine())
	log.Printf("[STATS] Memory Alloc: %d MB", m.Alloc/1024/1024)
	log.Printf("[STATS] Memory TotalAlloc: %d MB", m.TotalAlloc/1024/1024)
	log.Printf("[STATS] Memory Sys: %d MB", m.Sys/1024/1024)
	log.Printf("[STATS] NumGC: %d", m.NumGC)

	// Check available disk space
	dfCmd := exec.Command("df", "-h", "/tmp")
	dfOut, err := dfCmd.Output()
	if err == nil {
		lines := strings.Split(string(dfOut), "\n")
		if len(lines) > 1 {
			log.Printf("[STATS] Disk space: %s", strings.Join(strings.Fields(lines[1]), " "))
		}
	}
}

type CommitStats struct {
	Hash              string    `json:"hash"`
	Author            string    `json:"author"`
	Date              time.Time `json:"date"`
	Added             int       `json:"added"`
	Removed           int       `json:"removed"`
	Message           string    `json:"message"`
	FilesTouchedCount int       `json:"filesTouchedCount"`
}

type FileTouchCount struct {
	File  string `json:"file"`
	Count int    `json:"count"`
}

func getTopTouchedFiles(fileCounts map[string]int, limit int) []FileTouchCount {
	type fileCountPair struct {
		file  string
		count int
	}

	pairs := make([]fileCountPair, 0, len(fileCounts))
	for file, count := range fileCounts {
		pairs = append(pairs, fileCountPair{file: file, count: count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	if limit > len(pairs) {
		limit = len(pairs)
	}

	result := make([]FileTouchCount, limit)
	for i := 0; i < limit; i++ {
		result[i] = FileTouchCount{
			File:  pairs[i].file,
			Count: pairs[i].count,
		}
	}

	return result
}

func cloneRepo(repoURL string) ([]CommitStats, map[string]int, error) {
	overallStart := time.Now()

	// Log network test before clone
	log.Printf("[DEBUG] Testing network to GitHub before clone...")
	pingStart := time.Now()
	pingCmd := exec.Command("ping", "-c", "1", "github.com")
	pingOut, pingErr := pingCmd.CombinedOutput()
	if pingErr == nil {
		log.Printf("[DEBUG] Ping result: %s (took %v)", strings.TrimSpace(string(pingOut)), time.Since(pingStart))
	} else {
		log.Printf("[DEBUG] Ping failed: %v", pingErr)
	}

	tmpDirStart := time.Now()
	tmpDir, err := os.MkdirTemp("", "repo-analysis-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(tmpDir)
	log.Printf("[TIMING] Create temp dir: %v (path: %s)", time.Since(tmpDirStart), tmpDir)

	// Git clone
	cloneStart := time.Now()
	log.Printf("[DEBUG] Starting git clone for %s", repoURL)
	cloneCmd := exec.Command("git", "clone", "--bare", "--single-branch", repoURL, tmpDir)
	var cloneStderr bytes.Buffer
	cloneCmd.Stderr = &cloneStderr

	if err := cloneCmd.Run(); err != nil {
		log.Printf("Git clone failed for %s: %v - stderr: %s", repoURL, err, cloneStderr.String())
		return nil, nil, err
	}
	cloneDuration := time.Since(cloneStart)
	log.Printf("[TIMING] *** Git clone completed: %v ***", cloneDuration)

	// Check size of cloned repo
	duCmd := exec.Command("du", "-sh", tmpDir)
	duOut, _ := duCmd.Output()
	log.Printf("[DEBUG] Cloned repo size: %s", strings.TrimSpace(string(duOut)))

	// Git log
	gitLogStart := time.Now()
	log.Printf("[DEBUG] Starting git log for %s", repoURL)
	cmd := exec.Command("git",
		"--git-dir", tmpDir,
		"log",
		"--numstat",
		"--diff-algorithm=histogram",
		"--pretty=format:COMMIT:%H|%an|%at|%s",
	)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Git log failed for %s: %v ", repoURL, err)
		return nil, nil, err
	}
	gitLogDuration := time.Since(gitLogStart)
	log.Printf("[TIMING] *** Git log completed: %v ***", gitLogDuration)
	log.Printf("[DEBUG] Git log output size: %d bytes (%.2f MB)", len(output), float64(len(output))/1024/1024)

	// Parsing
	parseStart := time.Now()
	log.Printf("[DEBUG] Starting to parse git log output...")

	var commits []CommitStats
	fileTouchCounts := make(map[string]int)
	lines := strings.Split(string(output), "\n")
	log.Printf("[DEBUG] Total lines to parse: %d", len(lines))

	lineParseStart := time.Now()
	var currentCommit *CommitStats
	commitCount := 0

	for i, line := range lines {
		if i > 0 && i%10000 == 0 {
			log.Printf("[DEBUG] Parsed %d lines (%.1f%%), %d commits so far... (elapsed: %v)",
				i, float64(i)/float64(len(lines))*100, commitCount, time.Since(lineParseStart))
		}

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "COMMIT:") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
				commitCount++
			}

			commitLine := strings.TrimPrefix(line, "COMMIT:")
			parts := strings.SplitN(commitLine, "|", 4)
			if len(parts) == 4 {
				timestamp, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					timestamp = time.Now().Unix()
				}
				date := time.Unix(timestamp, 0)
				currentCommit = &CommitStats{
					Hash:              parts[0],
					Author:            parts[1],
					Date:              date,
					Message:           parts[3],
					Added:             0,
					Removed:           0,
					FilesTouchedCount: 0,
				}
			}
		} else if currentCommit != nil {
			tabFields := strings.Split(line, "\t")

			if len(tabFields) >= 3 {
				addedStr := tabFields[0]
				removedStr := tabFields[1]
				fileName := tabFields[2]

				added := 0
				removed := 0
				if addedStr != "-" {
					if parsed, err := strconv.Atoi(addedStr); err == nil {
						added = parsed
					}
				}
				if removedStr != "-" {
					if parsed, err := strconv.Atoi(removedStr); err == nil {
						removed = parsed
					}
				}

				currentCommit.Added += added
				currentCommit.Removed += removed

				if fileName != "" {
					currentCommit.FilesTouchedCount++
					fileTouchCounts[fileName]++
				}
			}
		}
	}

	if currentCommit != nil {
		commits = append(commits, *currentCommit)
		commitCount++
	}

	parseDuration := time.Since(parseStart)
	log.Printf("[TIMING] *** Parsing completed: %v ***", parseDuration)
	log.Printf("[DEBUG] Parsed %d commits, %d unique files", len(commits), len(fileTouchCounts))

	log.Printf("[TIMING] === CLONE REPO BREAKDOWN ===")
	log.Printf("[TIMING] - Temp dir creation: included in overall")
	log.Printf("[TIMING] - Git clone: %v (%.1f%%)", cloneDuration, float64(cloneDuration)/float64(time.Since(overallStart))*100)
	log.Printf("[TIMING] - Git log: %v (%.1f%%)", gitLogDuration, float64(gitLogDuration)/float64(time.Since(overallStart))*100)
	log.Printf("[TIMING] - Parsing: %v (%.1f%%)", parseDuration, float64(parseDuration)/float64(time.Since(overallStart))*100)
	log.Printf("[TIMING] - Total cloneRepo: %v", time.Since(overallStart))
	log.Printf("[TIMING] ================================")

	return commits, fileTouchCounts, nil
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
