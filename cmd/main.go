package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Yoosu-L/llmapibenchmark/internal/api"
	"github.com/Yoosu-L/llmapibenchmark/internal/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/pflag"
)

func main() {
	baseURL := pflag.StringP("base-url", "u", "", "Base URL of the OpenAI API")
	apiKey := pflag.StringP("api-key", "k", "", "API key for authentication")
	model := pflag.StringP("model", "m", "", "Model to be used for the requests (optional)")
	prompt := pflag.StringP("prompt", "p", "Write a long story, no less than 10,000 words, starting from a long, long time ago.", "Prompt to be used for generating responses")
	numWords := pflag.IntP("num-words", "n", 0, "Number of words Input")
	concurrencyStr := pflag.StringP("concurrency", "c", "1,2,4,8,16,32,64,128", "Comma-separated list of concurrency levels")
	maxTokens := pflag.IntP("max-tokens", "t", 512, "Maximum number of tokens to generate")
	help := pflag.BoolP("help", "h", false, "Show this help message")
	pflag.Parse()

	if *help {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// Parse concurrency levels
	concurrencyLevels, err := utils.ParseConcurrencyLevels(*concurrencyStr)
	if err != nil {
		log.Fatalf("Invalid concurrency levels: %v", err)
	}

	// Initialize OpenAI client
	var modelName string
	config := openai.DefaultConfig(*apiKey)
	config.BaseURL = *baseURL
	client := openai.NewClientWithConfig(config)

	// Discover model name if not provided
	if *model == "" {
		discoveredModel, err := api.GetFirstAvailableModel(client)
		if err != nil {
			log.Printf("Error discovering model: %v", err)
			return
		}
		modelName = discoveredModel
	} else {
		modelName = *model
	}

	// Determine input parameters and call benchmark function
	var inputTokens int
	var useRandomInput bool

	if *prompt != "Write a long story, no less than 10,000 words, starting from a long, long time ago." {
		useRandomInput = false
	} else if *numWords != 0 {
		useRandomInput = true
	} else {
		useRandomInput = false
	}

	// Get input tokens
	if useRandomInput {
		resp, err := api.AskOpenAIwithRandomInputNonStream(client, modelName, *numWords/4, 4)
		if err != nil {
			log.Fatalf("Error getting prompt tokens: %v", err)
		}
		inputTokens = resp.Usage.PromptTokens
	} else {
		resp, err := api.AskOpenAINonStream(client, modelName, *prompt, 4)
		if err != nil {
			log.Fatalf("Error getting prompt tokens: %v", err)
		}
		inputTokens = resp.Usage.PromptTokens
	}

	runBenchmark(*baseURL, *apiKey, modelName, *prompt, inputTokens, *maxTokens, concurrencyLevels, useRandomInput, *numWords)
}

func runBenchmark(baseURL, apiKey, modelName, prompt string, inputTokens, maxTokens int, concurrencyLevels []int, useRandomInput bool, numWords int) {
	// Test latency
	latency, err := utils.TestSpeedWithSystemProxy(baseURL, 5)
	if err != nil {
		log.Printf("Latency test error: %v", err)
		latency = 0
	}

	// Print benchmark header
	utils.PrintBenchmarkHeader(modelName, inputTokens, maxTokens, latency)

	// Print table header
	fmt.Println("| Concurrency | Generation Throughput (tokens/s) |  Prompt Throughput (tokens/s) | Min TTFT (s) | Max TTFT (s) |")
	fmt.Println("|-------------|----------------------------------|-------------------------------|--------------|--------------|")

	// Test each concurrency level and print results
	var results [][]interface{}
	for _, concurrency := range concurrencyLevels {
		var generationSpeed, promptThroughput, maxTTFT, minTTFT float64
		if useRandomInput {
			generationSpeed, promptThroughput, maxTTFT, minTTFT = utils.MeasureSpeedwithRandomInput(baseURL, apiKey, modelName, numWords/4, concurrency, inputTokens, maxTokens, latency)
		} else {
			generationSpeed, promptThroughput, maxTTFT, minTTFT = utils.MeasureSpeed(baseURL, apiKey, modelName, prompt, concurrency, inputTokens, maxTokens, latency)
		}

		// Print current results
		fmt.Printf("| %11d | %32.2f | %29.2f | %12.2f | %12.2f |\n",
			concurrency,
			generationSpeed,
			promptThroughput,
			minTTFT,
			maxTTFT)

		// Save results for later
		results = append(results, []interface{}{concurrency, generationSpeed, promptThroughput, minTTFT, maxTTFT})
	}

	fmt.Println("\n================================================================================================================")

	// Save results to Markdown
	utils.SaveResultsToMD(results, modelName, inputTokens, maxTokens, latency)
}
