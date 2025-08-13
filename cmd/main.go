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
	format := pflag.StringP("format", "f", "", "Output format")
	help := pflag.BoolP("help", "h", false, "Show this help message")
	pflag.Parse()

	if *help {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	benchmarkSetup := BenchmarkSetup{}
	benchmarkSetup.BaseURL = *baseURL
	benchmarkSetup.ApiKey = *apiKey
	benchmarkSetup.ModelName = *model
	benchmarkSetup.Prompt = *prompt
	benchmarkSetup.NumWords = *numWords
	benchmarkSetup.MaxTokens = *maxTokens

	// Parse concurrency levels
	concurrencyLevels, err := utils.ParseConcurrencyLevels(*concurrencyStr)
	if err != nil {
		log.Fatalf("Invalid concurrency levels: %v", err)
	}
	benchmarkSetup.ConcurrencyLevels = concurrencyLevels

	// Initialize OpenAI client
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
		benchmarkSetup.ModelName = discoveredModel
	}

	// Determine input parameters and call benchmark function
	if *prompt != "Write a long story, no less than 10,000 words, starting from a long, long time ago." {
		benchmarkSetup.UseRandomInput = false
	} else if *numWords != 0 {
		benchmarkSetup.UseRandomInput = true
	} else {
		benchmarkSetup.UseRandomInput = false
	}

	// Get input tokens
	if benchmarkSetup.UseRandomInput {
		resp, err := api.AskOpenAIwithRandomInputNonStream(client, benchmarkSetup.ModelName, *numWords/4, 4)
		if err != nil {
			log.Fatalf("Error getting prompt tokens: %v", err)
		}
		benchmarkSetup.InputTokens = resp.Usage.PromptTokens
	} else {
		resp, err := api.AskOpenAINonStream(client, benchmarkSetup.ModelName, *prompt, 4)
		if err != nil {
			log.Fatalf("Error getting prompt tokens: %v", err)
		}
		benchmarkSetup.InputTokens = resp.Usage.PromptTokens
	}

	if *format == "" {
		err := benchmarkSetup.runBenchmarkCli()
		if err != nil {
			log.Fatalf("Error running benchmark: %v", err)
		}
	} else {
		result, err := benchmarkSetup.runBenchmark()
		if err != nil {
			log.Fatalf("Error running benchmark: %v", err)
		}

		var output string
		switch *format {
		case "json":
			output, err = result.Json()
		case "yaml":
			output, err = result.Yaml()
		default:
			output, err = result.Json()
		}
		if err != nil {
			log.Fatalf("Error formatting benchmark: %v", err)
		}
		fmt.Println(output)
	}
}
