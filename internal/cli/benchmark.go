package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var (
	benchmarkModel       string
	benchmarkRequests    int
	benchmarkConcurrency int
	benchmarkProvider    string
	benchmarkLatestOnly  bool
	benchmarkTestURL     string
)

type BenchmarkReport struct {
	Model            string  `json:"model"`
	Requests         int     `json:"requests"`
	Concurrency      int     `json:"concurrency"`
	Provider         string  `json:"provider"`
	P50LatencyMs     int64   `json:"p50_latency_ms"`
	P95LatencyMs     int64   `json:"p95_latency_ms"`
	P99LatencyMs     int64   `json:"p99_latency_ms"`
	ThroughputTokens float64 `json:"throughput_tokens_sec"`
	RequestRate      float64 `json:"request_rate_sec"`
	ErrorRate        float64 `json:"error_rate"`
	TotalDurationSec float64 `json:"total_duration_sec"`
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark LLM inference workloads",
	Long:  `Run high-concurrency request benchmarks on inference workloads to measure throughput, latency, token rates, and cost efficiency.`,
}

var benchmarkRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an inference latency and throughput benchmark",
	Long:  `Sends concurrent request batches to the inference gateway and monitors response metrics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("🚀 %sRunning inference benchmark against model: %s%s%s...\n", utils.ColorBold, utils.ColorCyan, benchmarkModel, utils.ColorReset)
		fmt.Printf("%s[INFO] Target Provider: %s\n", utils.PrefixInfo, benchmarkProvider)
		fmt.Printf("%s[INFO] Concurrency: %d, Total Requests: %d\n", utils.PrefixInfo, benchmarkConcurrency, benchmarkRequests)
		fmt.Printf("%sDetecting target endpoint...\n", utils.PrefixInfo)

		// 1. Detect target URL
		var targetURL string
		var useGateway bool
		client := &http.Client{Timeout: 1 * time.Second}

		if benchmarkTestURL != "" {
			targetURL = benchmarkTestURL
			// For testing, treat it as gateway to exercise OAI path or direct Ollama
			if strings.Contains(benchmarkTestURL, "gateway") {
				useGateway = true
			}
		} else {
			if resp, err := client.Get("http://localhost:8000/models"); err == nil {
				_ = resp.Body.Close()
				targetURL = "http://localhost:8000/v1/chat/completions"
				useGateway = true
				fmt.Printf("%sSelected CloudInferOps Gateway at http://localhost:8000\n", utils.PrefixOK)
			} else if resp, err := client.Get("http://localhost:11434/api/tags"); err == nil {
				_ = resp.Body.Close()
				targetURL = "http://localhost:11434/api/generate"
				fmt.Printf("%sSelected local Ollama daemon at http://localhost:11434\n", utils.PrefixOK)
			} else {
				return fmt.Errorf("no active inference gateway (http://localhost:8000) or local Ollama daemon (http://localhost:11434) detected. Please make sure your inference stack is running")
			}
		}

		fmt.Printf("%sBenchmark run in progress...\n", utils.PrefixInfo)

		type ReqResult struct {
			Latency time.Duration
			Tokens  int
			Failed  bool
		}

		resultsChan := make(chan ReqResult, benchmarkRequests)
		jobsChan := make(chan int, benchmarkRequests)

		// 2. Start concurrent workers
		var wg sync.WaitGroup
		httpClient := &http.Client{Timeout: 30 * time.Second}

		for w := 1; w <= benchmarkConcurrency; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range jobsChan {
					start := time.Now()
					var tokens int
					var failed bool

					var req *http.Request
					var err error

					if useGateway {
						payload := fmt.Sprintf(`{"model": "%s", "messages": [{"role": "user", "content": "State a short fact about computer science."}], "stream": false}`, benchmarkModel)
						req, err = http.NewRequest("POST", targetURL, strings.NewReader(payload))
					} else {
						payload := fmt.Sprintf(`{"model": "%s", "prompt": "State a short fact about computer science.", "stream": false}`, benchmarkModel)
						req, err = http.NewRequest("POST", targetURL, strings.NewReader(payload))
					}

					if err == nil {
						req.Header.Set("Content-Type", "application/json")
						resp, err := httpClient.Do(req)
						if err == nil {
							var respBody bytes.Buffer
							_, _ = io.Copy(&respBody, resp.Body)
							_ = resp.Body.Close()

							if resp.StatusCode == http.StatusOK {
								if useGateway {
									var oaiResp struct {
										Usage struct {
											CompletionTokens int `json:"completion_tokens"`
										} `json:"usage"`
									}
									if json.Unmarshal(respBody.Bytes(), &oaiResp) == nil && oaiResp.Usage.CompletionTokens > 0 {
										tokens = oaiResp.Usage.CompletionTokens
									} else {
										tokens = len(respBody.String()) / 4
									}
								} else {
									var ollamaResp struct {
										EvalCount int `json:"eval_count"`
									}
									if json.Unmarshal(respBody.Bytes(), &ollamaResp) == nil && ollamaResp.EvalCount > 0 {
										tokens = ollamaResp.EvalCount
									} else {
										tokens = len(respBody.String()) / 4
									}
								}
							} else {
								failed = true
							}
						} else {
							failed = true
						}
					} else {
						failed = true
					}

					resultsChan <- ReqResult{
						Latency: time.Since(start),
						Tokens:  tokens,
						Failed:  failed,
					}
				}
			}()
		}

		// 3. Enqueue jobs and collect results
		go func() {
			for i := 0; i < benchmarkRequests; i++ {
				jobsChan <- i
			}
			close(jobsChan)
		}()

		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		var latencies []time.Duration
		totalTokens := 0
		failedRequests := 0
		startTime := time.Now()

		count := 0
		for res := range resultsChan {
			count++
			if res.Failed {
				failedRequests++
			} else {
				latencies = append(latencies, res.Latency)
				totalTokens += res.Tokens
			}
			percent := float64(count) / float64(benchmarkRequests) * 100
			fmt.Printf("\r\033[K%sProgress: %d/%d requests (%.1f%%)", utils.PrefixInfo, count, benchmarkRequests, percent)
		}
		fmt.Println()

		totalDuration := time.Since(startTime)
		fmt.Printf("%sBenchmark completed successfully.\n\n", utils.PrefixOK)

		// 4. Calculate percentiles and throughput metrics
		var p50, p95, p99 time.Duration
		if len(latencies) > 0 {
			sort.Slice(latencies, func(i, j int) bool {
				return latencies[i] < latencies[j]
			})
			safePercentile := func(pct float64) time.Duration {
				idx := int(float64(len(latencies)) * pct)
				if idx >= len(latencies) {
					idx = len(latencies) - 1
				}
				return latencies[idx]
			}
			p50 = safePercentile(0.50)
			p95 = safePercentile(0.95)
			p99 = safePercentile(0.99)
		}

		totalDurationSec := totalDuration.Seconds()
		if totalDurationSec <= 0 {
			totalDurationSec = 0.001
		}
		throughput := float64(totalTokens) / totalDurationSec
		reqRate := float64(benchmarkRequests) / totalDurationSec
		errorRate := float64(failedRequests) / float64(benchmarkRequests) * 100

		// Print summary results
		fmt.Printf("%s📊  Benchmark Results:%s\n", utils.ColorBold, utils.ColorReset)
		fmt.Println("--------------------------------------------------")
		fmt.Printf("    ⏱️  P50 Latency:       %v\n", p50)
		fmt.Printf("    ⏱️  P95 Latency:       %v\n", p95)
		fmt.Printf("    ⏱️  P99 Latency:       %v\n", p99)
		fmt.Printf("    ⚡  Token Throughput:  %.1f tokens/sec\n", throughput)
		fmt.Printf("    📈  Request Rate:      %.1f req/sec\n", reqRate)
		fmt.Printf("    🛑  Error Rate:        %.1f%%\n", errorRate)
		fmt.Println("--------------------------------------------------")

		// 5. Generate JSON report
		report := BenchmarkReport{
			Model:            benchmarkModel,
			Requests:         benchmarkRequests,
			Concurrency:      benchmarkConcurrency,
			Provider:         benchmarkProvider,
			P50LatencyMs:     p50.Milliseconds(),
			P95LatencyMs:     p95.Milliseconds(),
			P99LatencyMs:     p99.Milliseconds(),
			ThroughputTokens: throughput,
			RequestRate:      reqRate,
			ErrorRate:        errorRate,
			TotalDurationSec: totalDurationSec,
		}

		reportData, err := json.MarshalIndent(report, "", "  ")
		if err == nil {
			_ = os.MkdirAll("benchmarking/reports", 0755)
			_ = os.WriteFile("benchmarking/reports/latest.json", reportData, 0644)

			timestamp := time.Now().Format("20060102-150405")
			reportFile := fmt.Sprintf("benchmarking/reports/report-%s.json", timestamp)
			_ = os.WriteFile(reportFile, reportData, 0644)
			fmt.Printf("%sReport saved to benchmarking/reports/latest.json\n\n", utils.PrefixInfo)
		}

		return nil
	},
}

var benchmarkReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate or view benchmark reports",
	Long:  `Loads saved benchmark run outputs and presents details of throughput, latency, and estimated cost.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportFile := "benchmarking/reports/latest.json"
		reportData, err := os.ReadFile(reportFile)
		if err != nil {
			return fmt.Errorf("no benchmark report found at %s. Run a benchmark first using 'cloudinferops benchmark run'", reportFile)
		}

		var report BenchmarkReport
		if err := json.Unmarshal(reportData, &report); err != nil {
			return fmt.Errorf("failed to parse benchmark report: %w", err)
		}

		fmt.Printf("%sUnified Benchmark Report:%s\n", utils.ColorBold, utils.ColorReset)
		fmt.Println("==================================================")
		fmt.Printf("    Model:                 %s\n", report.Model)
		fmt.Printf("    Provider:              %s\n", report.Provider)
		fmt.Printf("    Requests:              %d\n", report.Requests)
		fmt.Printf("    Concurrency:           %d\n", report.Concurrency)
		fmt.Printf("    P50 Latency:           %dms\n", report.P50LatencyMs)
		fmt.Printf("    P95 Latency:           %dms\n", report.P95LatencyMs)
		fmt.Printf("    P99 Latency:           %dms\n", report.P99LatencyMs)
		fmt.Printf("    Tokens/sec:            %.1f\n", report.ThroughputTokens)
		fmt.Printf("    Request Rate:          %.1f req/sec\n", report.RequestRate)
		fmt.Printf("    Error Rate:            %.1f%%\n", report.ErrorRate)
		fmt.Printf("    Total Time:            %.1fs\n", report.TotalDurationSec)
		fmt.Println("==================================================")
		return nil
	},
}

func init() {
	benchmarkRunCmd.Flags().StringVar(&benchmarkModel, "model", "llama3", "Model to benchmark")
	benchmarkRunCmd.Flags().IntVar(&benchmarkRequests, "requests", 100, "Total requests to execute")
	benchmarkRunCmd.Flags().IntVar(&benchmarkConcurrency, "concurrency", 10, "Number of concurrent workers")
	benchmarkRunCmd.Flags().StringVar(&benchmarkProvider, "provider", "ollama", "Inference provider (ollama, vllm)")
	benchmarkRunCmd.Flags().StringVar(&benchmarkTestURL, "test-url", "", "Custom target URL for unit testing (hidden)")
	_ = benchmarkRunCmd.Flags().MarkHidden("test-url")

	benchmarkReportCmd.Flags().BoolVar(&benchmarkLatestOnly, "latest", true, "Show the latest benchmark report details")

	benchmarkCmd.AddCommand(benchmarkRunCmd)
	benchmarkCmd.AddCommand(benchmarkReportCmd)
	RootCmd.AddCommand(benchmarkCmd)
}
