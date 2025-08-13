package main

import (
	"fmt"

	"github.com/Yoosu-L/llmapibenchmark/internal/utils"
)

func (setup *BenchmarkSetup) runBenchmarkCli() error {
	// Test latency
	latency, err := utils.TestSpeedWithSystemProxy(setup.BaseURL, 5)
	if err != nil {
		return fmt.Errorf("latency test error: %v", err)
	}

	// Print benchmark header
	utils.PrintBenchmarkHeader(setup.ModelName, setup.InputTokens, setup.MaxTokens, latency)

	// Print table header
	fmt.Println("| Concurrency | Generation Throughput (tokens/s) |  Prompt Throughput (tokens/s) | Min TTFT (s) | Max TTFT (s) |")
	fmt.Println("|-------------|----------------------------------|-------------------------------|--------------|--------------|")

	// Test each concurrency level and print results
	var results [][]interface{}
	for _, concurrency := range setup.ConcurrencyLevels {
		measurement, err := setup.measureConcurrency(latency, concurrency)
		if err != nil {
			return fmt.Errorf("concurrency %d measurement error: %v", concurrency, err)
		}

		// Print current results
		fmt.Printf("| %11d | %32.2f | %29.2f | %12.2f | %12.2f |\n",
			concurrency,
			measurement.GenerationSpeed,
			measurement.PromptThroughput,
			measurement.MaxTtft,
			measurement.MaxTtft,
		)

		// Save results for later
		results = append(results, []interface{}{
			concurrency,
			measurement.GenerationSpeed,
			measurement.PromptThroughput,
			measurement.MaxTtft,
			measurement.MaxTtft,
		})
	}

	fmt.Println("\n================================================================================================================")

	// Save results to Markdown
	utils.SaveResultsToMD(results, setup.ModelName, setup.InputTokens, setup.MaxTokens, latency)

	return nil
}

func (setup *BenchmarkSetup) runBenchmark() (Benchmark, error) {
	benchmark := Benchmark{}

	// Test latency
	latency, err := utils.TestSpeedWithSystemProxy(setup.BaseURL, 5)
	if err != nil {
		return benchmark, fmt.Errorf("error testing latency: %v", err)
	}
	benchmark.Latency = latency

	for _, concurrency := range setup.ConcurrencyLevels {
		measurement, err := setup.measureConcurrency(latency, concurrency)
		if err != nil {
			return benchmark, fmt.Errorf("concurrency %d measurement error: %v", concurrency, err)
		}

		benchmark.Measurements = append(benchmark.Measurements, measurement)
	}

	return benchmark, nil
}

func (setup *BenchmarkSetup) measureConcurrency(latency float64, concurrency int) (utils.Measurement, error) {
	measurementSetup := utils.MeasurementSetup{
		BaseUrl:     setup.BaseURL,
		ApiKey:      setup.ApiKey,
		ModelName:   setup.ModelName,
		Prompt:      setup.Prompt,
		NumWords:    setup.NumWords,
		MaxTokens:   setup.MaxTokens,
		Latency:     latency,
		Concurrency: concurrency,
	}
	if setup.UseRandomInput {
		measurementSetup.UseRandomInput = true
	}

	var measurement utils.Measurement
	measurement = measurementSetup.MeasureSpeed()
	return measurement, nil
}
