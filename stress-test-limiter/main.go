package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestResult struct {
	StatusCode int
	Error      error
	Duration   time.Duration
}

type LoadTester struct {
	client      HTTPClient
	url         string
	totalReqs   int
	concurrency int
}

type TestReport struct {
	ExecutionTime time.Duration
	TotalRequests int
	SuccessCount  int
	StatusMap     map[int]int
}

type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

func main() {
	config := parseFlags()
	if !config.isValid() {
		displayUsage()
		return
	}

	tester := NewLoadTester(config)
	tester.displayTestInfo()

	report := tester.ExecuteTest()
	report.Display()
}

func parseFlags() Config {
	url := flag.String("url", "", "URL of the service to test")
	requests := flag.Int("requests", 0, "Total number of requests")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent requests")
	flag.Parse()

	return Config{
		URL:         *url,
		Requests:    *requests,
		Concurrency: *concurrency,
	}
}

func (c Config) isValid() bool {
	return c.URL != "" && c.Requests > 0 && c.Concurrency > 0
}

func displayUsage() {
	fmt.Println("Usage: stress-test --url=<URL> --requests=<NUM> --concurrency=<NUM>")
	fmt.Println("Example: stress-test --url=http://google.com --requests=1000 --concurrency=10")
}

func NewLoadTester(config Config) *LoadTester {
	return &LoadTester{
		client:      &http.Client{Timeout: 30 * time.Second},
		url:         config.URL,
		totalReqs:   config.Requests,
		concurrency: config.Concurrency,
	}
}

func (lt *LoadTester) displayTestInfo() {
	fmt.Printf("🚀 Initiating Load Test\n")
	fmt.Printf("Target URL: %s\n", lt.url)
	fmt.Printf("Total Requests: %d\n", lt.totalReqs)
	fmt.Printf("Concurrent Workers: %d\n", lt.concurrency)
	fmt.Printf("----------------------------------------\n")
}

func (lt *LoadTester) ExecuteTest() TestReport {
	startTime := time.Now()

	jobs := make(chan int, lt.totalReqs)
	results := make(chan RequestResult, lt.totalReqs)

	var wg sync.WaitGroup
	for w := 0; w < lt.concurrency; w++ {
		wg.Add(1)
		go lt.worker(jobs, results, &wg)
	}

	for j := 0; j < lt.totalReqs; j++ {
		jobs <- j
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	statusMap := make(map[int]int)
	successCount := 0

	for result := range results {
		if result.Error != nil {
			statusMap[0]++
		} else {
			statusMap[result.StatusCode]++
			if result.StatusCode == 200 {
				successCount++
			}
		}
	}

	return TestReport{
		ExecutionTime: time.Since(startTime),
		TotalRequests: lt.totalReqs,
		SuccessCount:  successCount,
		StatusMap:     statusMap,
	}
}

func (lt *LoadTester) worker(jobs <-chan int, results chan<- RequestResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for range jobs {
		result := lt.performRequest()
		results <- result
	}
}

func (lt *LoadTester) performRequest() RequestResult {
	req, err := http.NewRequest("GET", lt.url, nil)
	if err != nil {
		return RequestResult{Error: err}
	}

	start := time.Now()
	resp, err := lt.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{Error: err, Duration: duration}
	}
	defer resp.Body.Close()

	return RequestResult{
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}
}

func (tr TestReport) Display() {
	fmt.Printf("\n📊 === LOAD TEST RESULTS ===\n")
	fmt.Printf("⏱️  Execution Time: %v\n", tr.ExecutionTime)
	fmt.Printf("📝 Total Requests: %d\n", tr.TotalRequests)
	fmt.Printf("✅ Successful (200 OK): %d\n", tr.SuccessCount)

	successRate := float64(tr.SuccessCount) / float64(tr.TotalRequests) * 100
	fmt.Printf("📈 Success Rate: %.2f%%\n", successRate)

	fmt.Printf("\n📋 Status Code Breakdown:\n")
	for status, count := range tr.StatusMap {
		percentage := float64(count) / float64(tr.TotalRequests) * 100

		if status == 0 {
			fmt.Printf("   ❌ Connection Errors: %d (%.1f%%)\n", count, percentage)
		} else {
			emoji := getStatusEmoji(status)
			fmt.Printf("   %s HTTP %d: %d (%.1f%%)\n", emoji, status, count, percentage)
		}
	}
	fmt.Printf("----------------------------------------\n")
}

func getStatusEmoji(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "✅"
	case status >= 300 && status < 400:
		return "🔄"
	case status >= 400 && status < 500:
		return "⚠️"
	case status >= 500:
		return "❌"
	default:
		return "❓"
	}
}
