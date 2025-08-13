package utils

import (
	"math"
	"sync"
	"time"

	"github.com/Yoosu-L/llmapibenchmark/internal/api"

	"github.com/sashabaranov/go-openai"
)

type MeasurementSetup struct {
	BaseUrl        string
	ApiKey         string
	ModelName      string
	Prompt         string
	UseRandomInput bool
	NumWords       int
	MaxTokens      int
	Latency        float64
	Concurrency    int
}

type Measurement struct {
	GenerationSpeed  float64 `json:"generation_speed"`
	PromptThroughput float64 `json:"prompt_throughput"`
	MaxTtft          float64 `json:"max_ttft"`
	MinTtft          float64 `json:"min_ttft"`
}

// MeasureSpeed measures API generation throughput and TTFT.
func (setup *MeasurementSetup) MeasureSpeed() Measurement {
	config := openai.DefaultConfig(setup.ApiKey)
	config.BaseURL = setup.BaseUrl
	client := openai.NewClientWithConfig(config)

	var wg sync.WaitGroup
	var responseTokens sync.Map
	var promptTokens sync.Map
	var ttfts sync.Map

	start := time.Now()

	// Send requests concurrently
	for i := 0; i < setup.Concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			var ttft float64
			var completionTokens, inputTokens int
			var err error
			if setup.UseRandomInput {
				ttft, completionTokens, inputTokens, err = api.AskOpenAIwithRandomInput(client, setup.ModelName, setup.NumWords, setup.MaxTokens)
			} else {
				ttft, completionTokens, inputTokens, err = api.AskOpenAI(client, setup.ModelName, setup.Prompt, setup.MaxTokens)
			}
			if err != nil {
				// TODO use a mutex to store err and handle in outer thread
				return
			}
			ttfts.Store(index, ttft)
			responseTokens.Store(index, completionTokens)
			promptTokens.Store(index, inputTokens)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate total tokens
	totalResponseTokens := 0
	responseTokens.Range(func(_, value interface{}) bool {
		totalResponseTokens += value.(int)
		return true
	})

	totalPromptTokens := 0
	promptTokens.Range(func(_, value interface{}) bool {
		totalPromptTokens += value.(int)
		return true
	})

	measurement := Measurement{}

	// Calculate max and min TTFT
	measurement.MaxTtft = 0.0
	measurement.MinTtft = math.Inf(1)
	ttfts.Range(func(_, value interface{}) bool {
		ttft := value.(float64)
		if ttft > measurement.MaxTtft {
			measurement.MaxTtft = ttft
		}
		if ttft < measurement.MinTtft {
			measurement.MinTtft = ttft
		}
		return true
	})

	// Calculate speed (tokens/second)
	measurement.GenerationSpeed = float64(totalResponseTokens) / (duration.Seconds() - setup.Latency/1000)

	// Calculate Prompt Throughput
	measurement.PromptThroughput = float64(totalPromptTokens) / (measurement.MaxTtft - setup.Latency/1000)

	return measurement
}
