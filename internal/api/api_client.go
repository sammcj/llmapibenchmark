package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AskOpenAI sends a prompt to the OpenAI API and retrieves the response.
func AskOpenAI(client *openai.Client, model, prompt string, maxTokens int) (float64, int, int, error) {
	start := time.Now()
	var ttft float64
	var completionTokens, promptTokens int

	stream, err := client.CreateChatCompletionStream(
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
			MaxTokens:   maxTokens,
			Temperature: 1,
			Stream:      true,
		},
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer stream.Close()

	first := true
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, 0, fmt.Errorf("stream error: %w", err)
		}
		if first {
			ttft = time.Since(start).Seconds()
			if response.Usage != nil {
				promptTokens = response.Usage.PromptTokens
			}
			first = false
		}
		if len(response.Choices) > 0 {
			completionTokens += len(response.Choices[0].Delta.Content)
		}
	}
	return ttft, completionTokens, promptTokens, nil
}

// AskOpenAIwithRandomInput sends a prompt to the OpenAI API and retrieves the response.
func AskOpenAIwithRandomInput(client *openai.Client, model string, numWords int, maxTokens int) (float64, int, int, error) {
	prompt := generateRandomPhrase(numWords)
	start := time.Now()
	var ttft float64
	var completionTokens, promptTokens int

	stream, err := client.CreateChatCompletionStream(
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
			MaxTokens:   maxTokens,
			Temperature: 1,
			Stream:      true,
		},
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer stream.Close()

	first := true
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, 0, fmt.Errorf("stream error: %w", err)
		}
		if first {
			ttft = time.Since(start).Seconds()
			if response.Usage != nil {
				promptTokens = response.Usage.PromptTokens
			}
			first = false
		}
		if len(response.Choices) > 0 {
			completionTokens += len(response.Choices[0].Delta.Content)
		}
	}
	return ttft, completionTokens, promptTokens, nil
}

// AskOpenAI with no stream
func AskOpenAINonStream(client *openai.Client, model, prompt string, maxTokens int) (*openai.ChatCompletionResponse, error) {
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
			MaxTokens:   maxTokens,
			Temperature: 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	return &resp, nil
}

// AskOpenAIwithRandomInput with no stream
func AskOpenAIwithRandomInputNonStream(client *openai.Client, model string, numWords int, maxTokens int) (*openai.ChatCompletionResponse, error) {
	prompt := generateRandomPhrase(numWords)
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
			MaxTokens:   maxTokens,
			Temperature: 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	return &resp, nil
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
