package main

import "github.com/Yoosu-L/llmapibenchmark/internal/utils"

type BenchmarkSetup struct {
	BaseURL           string `json:"base_url"`
	ApiKey            string `json:"api_key"`
	ModelName         string `json:"model_name"`
	Prompt            string `json:"prompt"`
	InputTokens       int    `json:"input_tokens"`
	MaxTokens         int    `json:"max_tokens"`
	ConcurrencyLevels []int  `json:"concurrency_levels"`
	UseRandomInput    bool   `json:"use_random_input"`
	NumWords          int    `json:"num_words"`
}

type Benchmark struct {
	ModelName    string
	InputTokens  int
	MaxTokens    int
	Latency      float64
	Measurements []utils.Measurement
}
