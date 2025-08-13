package main

import "github.com/Yoosu-L/llmapibenchmark/internal/utils"

type Benchmark struct {
	BaseURL           string
	ApiKey            string
	ModelName         string
	Prompt            string
	InputTokens       int
	MaxTokens         int
	ConcurrencyLevels []int
	UseRandomInput    bool
	NumWords          int
}

type BenchmarkResult struct {
	ModelName   string              `json:"model_name" yaml:"model-name"`
	InputTokens int                 `json:"input_tokens" yaml:"input-tokens"`
	MaxTokens   int                 `json:"max_tokens" yaml:"max-tokens"`
	Latency     float64             `json:"latency" yaml:"latency"`
	Results     []utils.SpeedResult `json:"results" yaml:"results"`
}
