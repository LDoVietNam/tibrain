package benchmark

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ti/cli/internal/core"
)

// Benchmark performs performance benchmarking on plugins
type Benchmark struct {
	mu       sync.RWMutex
	platform core.Platform
	results  map[string]*BenchmarkResult
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	PluginName    string
	Task          string
	TotalRuns     int
	SuccessCount  int
	FailureCount  int
	MinLatency    time.Duration
	MaxLatency    time.Duration
	AvgLatency    time.Duration
	P50Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	Throughput    float64 // requests per second
	ErrorRate     float64
	Timestamp     time.Time
	TotalDuration time.Duration
	latencies     []time.Duration
}

// Config holds benchmark configuration
type Config struct {
	PluginName string
	Task       string
	Input      map[string]interface{}
	Runs       int
	Concurrent int
}

// NewBenchmark creates a new benchmark
func NewBenchmark(platform core.Platform) *Benchmark {
	return &Benchmark{
		platform: platform,
		results:  make(map[string]*BenchmarkResult),
	}
}

// Run runs a benchmark on a plugin
func (b *Benchmark) Run(ctx context.Context, config Config) (*BenchmarkResult, error) {
	if config.Runs == 0 {
		config.Runs = 100
	}
	if config.Concurrent == 0 {
		config.Concurrent = 1
	}

	// Initialize result
	result := &BenchmarkResult{
		PluginName: config.PluginName,
		Task:       config.Task,
		TotalRuns:  config.Runs,
		Timestamp:  time.Now(),
	}

	// Run benchmark
	if config.Concurrent == 1 {
		err := b.runSequential(ctx, config, result)
		if err != nil {
			return nil, err
		}
	} else {
		err := b.runConcurrent(ctx, config, result)
		if err != nil {
			return nil, err
		}
	}

	// Calculate statistics
	b.calculateStatistics(result)

	// Store result
	b.mu.Lock()
	b.results[config.PluginName+"_"+config.Task] = result
	b.mu.Unlock()

	return result, nil
}

// runSequential runs benchmark sequentially
func (b *Benchmark) runSequential(ctx context.Context, config Config, result *BenchmarkResult) error {
	latencies := make([]time.Duration, 0, config.Runs)

	for i := 0; i < config.Runs; i++ {
		start := time.Now()

		_, err := b.platform.ExecutePlugin(ctx, config.PluginName, config.Task, config.Input)
		latency := time.Since(start)

		if err != nil {
			result.FailureCount++
		} else {
			result.SuccessCount++
			latencies = append(latencies, latency)
		}
	}

	result.latencies = latencies
	return nil
}

// runConcurrent runs benchmark concurrently
func (b *Benchmark) runConcurrent(ctx context.Context, config Config, result *BenchmarkResult) error {
	latencies := make([]time.Duration, 0, config.Runs)
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()

	for i := 0; i < config.Runs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			start := time.Now()
			_, err := b.platform.ExecutePlugin(ctx, config.PluginName, config.Task, config.Input)
			latency := time.Since(start)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailureCount++
			} else {
				result.SuccessCount++
				latencies = append(latencies, latency)
			}
		}()

		// Limit concurrency
		if (i+1)%config.Concurrent == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	result.TotalDuration = totalDuration
	result.latencies = latencies
	return nil
}

// calculateStatistics calculates benchmark statistics
func (b *Benchmark) calculateStatistics(result *BenchmarkResult) {
	latencies := result.latencies
	if len(latencies) == 0 {
		return
	}

	// Min/Max/Avg
	result.MinLatency = latencies[0]
	result.MaxLatency = latencies[0]
	var total time.Duration
	for _, lat := range latencies {
		if lat < result.MinLatency {
			result.MinLatency = lat
		}
		if lat > result.MaxLatency {
			result.MaxLatency = lat
		}
		total += lat
	}
	result.AvgLatency = total / time.Duration(len(latencies))

	// Percentiles
	result.P50Latency = b.percentile(latencies, 50)
	result.P95Latency = b.percentile(latencies, 95)
	result.P99Latency = b.percentile(latencies, 99)

	// Throughput
	if result.TotalDuration > 0 {
		result.Throughput = float64(result.SuccessCount) / result.TotalDuration.Seconds()
	}

	// Error rate
	result.ErrorRate = float64(result.FailureCount) / float64(result.TotalRuns)
}

// percentile calculates a percentile
func (b *Benchmark) percentile(latencies []time.Duration, p int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Simple sort (in production, use more efficient algorithm)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile index
	index := (p * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// GetResult returns a benchmark result
func (b *Benchmark) GetResult(pluginName, task string) (*BenchmarkResult, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result, exists := b.results[pluginName+"_"+task]
	return result, exists
}

// GetAllResults returns all benchmark results
func (b *Benchmark) GetAllResults() map[string]*BenchmarkResult {
	b.mu.RLock()
	defer b.mu.RUnlock()

	results := make(map[string]*BenchmarkResult)
	for k, v := range b.results {
		results[k] = v
	}
	return results
}

// ClearResults clears all benchmark results
func (b *Benchmark) ClearResults() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.results = make(map[string]*BenchmarkResult)
}

// PrintResult prints a benchmark result
func (b *Benchmark) PrintResult(result *BenchmarkResult) {
	fmt.Printf("\n=== Benchmark Results ===\n")
	fmt.Printf("Plugin: %s\n", result.PluginName)
	fmt.Printf("Task: %s\n", result.Task)
	fmt.Printf("Total Runs: %d\n", result.TotalRuns)
	fmt.Printf("Success: %d\n", result.SuccessCount)
	fmt.Printf("Failures: %d\n", result.FailureCount)
	fmt.Printf("Error Rate: %.2f%%\n", result.ErrorRate*100)
	fmt.Printf("\nLatency:\n")
	fmt.Printf("  Min: %v\n", result.MinLatency)
	fmt.Printf("  Max: %v\n", result.MaxLatency)
	fmt.Printf("  Avg: %v\n", result.AvgLatency)
	fmt.Printf("  P50: %v\n", result.P50Latency)
	fmt.Printf("  P95: %v\n", result.P95Latency)
	fmt.Printf("  P99: %v\n", result.P99Latency)
	fmt.Printf("\nThroughput: %.2f req/s\n", result.Throughput)
	fmt.Printf("Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	fmt.Printf("========================\n")
}
