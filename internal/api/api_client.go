package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/schollz/progressbar/v3"
)

// AskOpenAiStream sends a prompt to the OpenAI API, processes the response stream and returns stats on it.
func AskOpenAiStream(client *openai.Client, model string, prompt string, maxTokens int) (float64, int, int, error) {
	start := time.Now()

	var (
		timeToFirstToken float64
		firstTokenSeen   bool
		lastUsage        *openai.Usage
	)

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// Add the deprecated `MaxTokens` for backward compatibility with some older API servers.
			MaxTokens:           maxTokens,
			MaxCompletionTokens: maxTokens,
			Temperature:         1,
			Stream:              true,
			StreamOptions: &openai.StreamOptions{
				IncludeUsage: true,
			},
		},
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, 0, fmt.Errorf("stream error: %w", err)
		}

		if !firstTokenSeen && len(resp.Choices) > 0 {
			content := resp.Choices[0].Delta.Content
			if strings.TrimSpace(content) != "" {
				timeToFirstToken = time.Since(start).Seconds()
				firstTokenSeen = true
			}
		}

		if resp.Usage != nil {
			lastUsage = resp.Usage
		}
	}

	var promptTokens, completionTokens int
	if lastUsage != nil {
		promptTokens = lastUsage.PromptTokens
		completionTokens = lastUsage.CompletionTokens
	}

	return timeToFirstToken, completionTokens, promptTokens, nil
}

func AskOpenAiStreamWithRandomInput(client *openai.Client, model string, numWords int, maxTokens int) (float64, int, int, error) {
	prompt := generateRandomPhrase(numWords)
	return AskOpenAiStream(client, model, prompt, maxTokens)
}

// AskOpenAiStreamWithProgress sends a prompt to the OpenAI API with progress bar updates.
func AskOpenAiStreamWithProgress(client *openai.Client, model string, prompt string, maxTokens int, bar *progressbar.ProgressBar) (float64, int, int, error) {
	start := time.Now()

	var (
		timeToFirstToken float64
		firstTokenSeen   bool
		lastUsage        *openai.Usage
		accumulatedContent string // Accumulate all content to count tokens more accurately
		estimatedTokens  int      // Real-time token estimation
	)

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// Add the deprecated `MaxTokens` for backward compatibility with some older API servers.
			MaxTokens:           maxTokens,
			MaxCompletionTokens: maxTokens,
			Temperature:         1,
			Stream:              true,
			StreamOptions: &openai.StreamOptions{
				IncludeUsage: true,
			},
		},
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, 0, fmt.Errorf("stream error: %w", err)
		}

		if !firstTokenSeen && len(resp.Choices) > 0 {
			content := resp.Choices[0].Delta.Content
			if strings.TrimSpace(content) != "" {
				timeToFirstToken = time.Since(start).Seconds()
				firstTokenSeen = true
			}
		}

		// Process each chunk and estimate tokens in real-time
		if len(resp.Choices) > 0 {
			content := resp.Choices[0].Delta.Content
			if content != "" {
				accumulatedContent += content
				
				// Improved token estimation based on content characteristics
				newTokens := estimateTokensFromContent(content)
				estimatedTokens += newTokens
				
				if bar != nil {
					bar.Add(newTokens)
				}
			}
		}

		if resp.Usage != nil {
			lastUsage = resp.Usage
		}
	}

	var promptTokens, completionTokens int
	if lastUsage != nil {
		promptTokens = lastUsage.PromptTokens
		completionTokens = lastUsage.CompletionTokens
		
		// Final adjustment: if we have actual completion tokens, adjust the progress bar
		if bar != nil && completionTokens > 0 {
			diff := completionTokens - estimatedTokens
			if diff != 0 { // Could be positive or negative
				bar.Add(diff)
			}
		}
	} else {
		// If no usage info, use our estimated tokens as completion tokens
		completionTokens = estimatedTokens
	}

	return timeToFirstToken, completionTokens, promptTokens, nil
}

// estimateTokensFromContent provides more accurate token estimation based on content
func estimateTokensFromContent(content string) int {
	if content == "" {
		return 0
	}
	
	// More sophisticated token estimation algorithm
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return 0
	}
	
	// Different strategies based on content type
	words := strings.Fields(content)
	wordCount := len(words)
	
	if wordCount > 0 {
		// For text with clear word boundaries: ~1.3 tokens per word on average
		// This accounts for subword tokenization in modern models
		return max(1, int(float64(wordCount)*1.3))
	} else {
		// For content without clear word boundaries (like punctuation, single chars)
		// Use character-based estimation: ~3-4 characters per token
		charCount := len(content)
		return max(1, charCount/3)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for Go versions that might not have max built-in
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func AskOpenAiStreamWithRandomInputAndProgress(client *openai.Client, model string, numWords int, maxTokens int, bar *progressbar.ProgressBar) (float64, int, int, error) {
	prompt := generateRandomPhrase(numWords)
	return AskOpenAiStreamWithProgress(client, model, prompt, maxTokens, bar)
}

// AskOpenAi sends a prompt to the OpenAI API and returns the response, not using streaming.
func AskOpenAi(client *openai.Client, model, prompt string, maxTokens int) (*openai.ChatCompletionResponse, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// Add the deprecated `MaxTokens` for backward compatibility with some older API servers.
			MaxTokens:           maxTokens,
			MaxCompletionTokens: maxTokens,
			Temperature:         1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	return &resp, nil
}

func AskOpenAiWithRandomInput(client *openai.Client, model string, numWords int, maxTokens int) (*openai.ChatCompletionResponse, error) {
	prompt := generateRandomPhrase(numWords)
	return AskOpenAi(client, model, prompt, maxTokens)
}

// GetFirstAvailableModel retrieves the first available model from the OpenAI API.
func GetFirstAvailableModel(client *openai.Client) (string, error) {
	modelList, err := client.ListModels(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list models: %w", err)
	}

	if len(modelList.Models) == 0 {
		return "", fmt.Errorf("no models available")
	}

	return modelList.Models[0].ID, nil
}
