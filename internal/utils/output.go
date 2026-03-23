package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// PrintBenchmarkHeader prints the benchmark header with details about the test.
func PrintBenchmarkHeader(modelName string, inputTokens int, maxTokens int, latency float64) {
	width := 80
	title := "LLM API Throughput Benchmark"
	url := "https://github.com/Yoosu-L/llmapibenchmark"
	timeStr := fmt.Sprintf("Time: %s", time.Now().UTC().Format("2006-01-02 15:04:05 UTC+0"))

	// ANSI Colors
	cyan := "\033[36m"
	green := "\033[32m"
	dim := "\033[2m"
	bold := "\033[1m"
	reset := "\033[0m"

	border := strings.Repeat("#", width)

	center := func(s string, w int) string {
		if len(s) >= w {
			return s
		}
		left := (w - len(s)) / 2
		right := w - len(s) - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}

	fmt.Println(cyan + border + reset)
	fmt.Println(bold + center(title, width) + reset)
	fmt.Println(dim + center(url, width) + reset)
	fmt.Println(dim + center(timeStr, width) + reset)
	fmt.Println(cyan + border + reset)

	fmt.Printf("%s%sModel:%s %-25s | %sLatency:%s %.2f ms%s\n", green, bold, reset+green, modelName, bold, reset+green, latency, reset)
	fmt.Printf("%s%sInput:%s %-25d | %sOutput:%s  %d tokens%s\n\n", green, bold, reset+green, inputTokens, bold, reset+green, maxTokens, reset)
}

// SaveResultsToMD saves the benchmark results to a Markdown file.
func SaveResultsToMD(results [][]interface{}, modelName string, inputTokens int, maxTokens int, latency float64) {
	// sanitize modelName to create a safe filename (replace path separators)
	safeModelName := strings.ReplaceAll(modelName, "/", "_")
	safeModelName = strings.ReplaceAll(safeModelName, "\\", "_")
	safeModelName = strings.TrimSpace(safeModelName)
	if safeModelName == "" {
		safeModelName = "model"
	}
	filename := fmt.Sprintf("API_Throughput_%s.md", safeModelName)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("```\nModel: %s\nLatency: %.2f ms\nInput: %d tokens / Output: %d tokens\n```\n\n", modelName, latency, inputTokens, maxTokens))
	file.WriteString("| Conc | Gen TPS | Prompt TPS | Min TTFT(s) | Max TTFT(s) | Success | Total(s) |\n")
	file.WriteString("|:----:|:-------:|:----------:|:-----------:|:-----------:|:-------:|:--------:|\n")

	for _, result := range results {
		concurrency := result[0].(int)
		generationSpeed := result[1].(float64)
		promptThroughput := result[2].(float64)
		minTTFT := result[3].(float64)
		maxTTFT := result[4].(float64)
		successRate := result[5].(float64)
		duration := result[6].(float64)
		file.WriteString(fmt.Sprintf("| %4d | %7.2f | %10.2f | %11.2f | %11.2f | %6.2f%% | %8.2f |\n",
			concurrency,
			generationSpeed,
			promptThroughput,
			minTTFT,
			maxTTFT,
			successRate*100,
			duration,
		))
	}

	fmt.Printf("Results saved to: %s\n\n", filename)
}
