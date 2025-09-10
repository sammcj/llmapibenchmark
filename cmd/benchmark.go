package main

import (
	"fmt"

	"github.com/Yoosu-L/llmapibenchmark/internal/utils"
	"github.com/schollz/progressbar/v3"
)

func (benchmark *Benchmark) runCli() error {
	// Test latency
	latency, err := utils.MeasureLatency(benchmark.BaseURL, 5)
	if err != nil {
		return fmt.Errorf("latency test error: %v", err)
	}

	// Print benchmark header
	utils.PrintBenchmarkHeader(benchmark.ModelName, benchmark.InputTokens, benchmark.MaxTokens, latency)

	// Print table header
	fmt.Println("| Concurrency | Generation Throughput (tokens/s) |  Prompt Throughput (tokens/s) | Min TTFT (s) | Max TTFT (s) |")
	fmt.Println("|-------------|----------------------------------|-------------------------------|--------------|--------------|")

	// Test each concurrency level and print results
	var results [][]interface{}
	for _, concurrency := range benchmark.ConcurrencyLevels {
		// Create a progress bar for this specific concurrency level
		expectedTokens := concurrency * benchmark.MaxTokens
		bar := progressbar.NewOptions(expectedTokens,
			progressbar.OptionSetDescription(fmt.Sprintf("Concurrency %d", concurrency)),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("tokens"),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionSetRenderBlankState(true),
		)

		result, err := benchmark.measureSpeedWithProgress(latency, concurrency, bar)
		if err != nil {
			return fmt.Errorf("concurrency %d: %v", concurrency, err)
		}

		bar.Finish()
		bar.Clear() // Clear the progress bar from the terminal
		bar.Close() // Clean up the progress bar

		// Print current results
		fmt.Printf("| %11d | %32.2f | %29.2f | %12.2f | %12.2f |\n",
			concurrency,
			result.GenerationSpeed,
			result.PromptThroughput,
			result.MinTtft,
			result.MaxTtft,
		)

		// Save results for later
		results = append(results, []interface{}{
			concurrency,
			result.GenerationSpeed,
			result.PromptThroughput,
			result.MinTtft,
			result.MaxTtft,
		})
	}

	fmt.Println("\n================================================================================================================")

	// Save results to Markdown
	utils.SaveResultsToMD(results, benchmark.ModelName, benchmark.InputTokens, benchmark.MaxTokens, latency)

	return nil
}

func (benchmark *Benchmark) run() (BenchmarkResult, error) {
	result := BenchmarkResult{}
	result.ModelName = benchmark.ModelName
	result.InputTokens = benchmark.InputTokens
	result.MaxTokens = benchmark.MaxTokens

	// Test latency
	latency, err := utils.MeasureLatency(benchmark.BaseURL, 5)
	if err != nil {
		return result, fmt.Errorf("error testing latency: %v", err)
	}
	result.Latency = latency

	for _, concurrency := range benchmark.ConcurrencyLevels {
		measurement, err := benchmark.measureSpeed(latency, concurrency)
		if err != nil {
			return result, fmt.Errorf("concurrency %d: %v", concurrency, err)
		}

		result.Results = append(result.Results, measurement)
	}

	return result, nil
}

func (benchmark *Benchmark) measureSpeed(latency float64, concurrency int) (utils.SpeedResult, error) {
	speedMeasurement := utils.SpeedMeasurement{
		BaseUrl:     benchmark.BaseURL,
		ApiKey:      benchmark.ApiKey,
		ModelName:   benchmark.ModelName,
		Prompt:      benchmark.Prompt,
		NumWords:    benchmark.NumWords,
		MaxTokens:   benchmark.MaxTokens,
		Latency:     latency,
		Concurrency: concurrency,
	}
	if benchmark.UseRandomInput {
		speedMeasurement.UseRandomInput = true
	}

	var result utils.SpeedResult
	result, err := speedMeasurement.Run()
	if err != nil {
		return result, fmt.Errorf("measurement error: %v", err)
	}
	return result, nil
}

func (benchmark *Benchmark) measureSpeedWithProgress(latency float64, concurrency int, bar *progressbar.ProgressBar) (utils.SpeedResult, error) {
	speedMeasurement := utils.SpeedMeasurement{
		BaseUrl:     benchmark.BaseURL,
		ApiKey:      benchmark.ApiKey,
		ModelName:   benchmark.ModelName,
		Prompt:      benchmark.Prompt,
		NumWords:    benchmark.NumWords,
		MaxTokens:   benchmark.MaxTokens,
		Latency:     latency,
		Concurrency: concurrency,
	}
	if benchmark.UseRandomInput {
		speedMeasurement.UseRandomInput = true
	}

	var result utils.SpeedResult
	result, err := speedMeasurement.RunWithProgress(bar)
	if err != nil {
		return result, fmt.Errorf("measurement error: %v", err)
	}
	return result, nil
}
