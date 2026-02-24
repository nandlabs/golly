// Package main demonstrates the Claude (Anthropic) GenAI provider.
//
// Prerequisites:
//
//	export ANTHROPIC_API_KEY="sk-ant-..."
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"oss.nandlabs.io/golly/genai"
	"oss.nandlabs.io/golly/genai/impl"
)

const defaultModel = "claude-sonnet-4-20250514"

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create the Claude provider (uses x-api-key header auth automatically)
	provider := impl.NewClaudeProvider(apiKey, nil)
	defer provider.Close()

	fmt.Printf("Provider: %s (v%s)\n", provider.Name(), provider.Version())
	fmt.Printf("Description: %s\n\n", provider.Description())

	// --- Basic text generation ---
	fmt.Println("=== Basic Generation ===")
	basicGeneration(provider)

	// --- Generation with system instructions ---
	fmt.Println("\n=== System Instructions ===")
	systemInstructions(provider)

	// --- Streaming generation ---
	fmt.Println("\n=== Streaming ===")
	streaming(provider)

	// --- Using top_k (Claude-specific) ---
	fmt.Println("\n=== Top-K Sampling (Claude-specific) ===")
	topKSampling(provider)
}

func basicGeneration(provider genai.Provider) {
	msg := genai.NewTextMessage(genai.RoleUser, "What are the three primary colors? Answer briefly.")
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(128).
		SetTemperature(0.3).
		Build()

	resp, err := provider.Generate(context.Background(), defaultModel, msg, opts)
	if err != nil {
		log.Printf("Generation error: %v\n", err)
		return
	}

	printResponse(resp)
	fmt.Printf("Tokens: input=%d, output=%d, total=%d\n",
		resp.Meta.InputTokens, resp.Meta.OutputTokens, resp.Meta.TotalTokens)
}

func systemInstructions(provider genai.Provider) {
	// Claude handles system instructions as a top-level "system" field,
	// not as a message. The provider maps this automatically.
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetTemperature(0.7).
		Build()
	opts.Set(genai.OptionSystemInstructions,
		"You are a Shakespearean actor. Respond in iambic pentameter. Keep it to 4 lines.")

	msg := genai.NewTextMessage(genai.RoleUser, "What is your opinion on modern technology?")
	resp, err := provider.Generate(context.Background(), defaultModel, msg, opts)
	if err != nil {
		log.Printf("Generation error: %v\n", err)
		return
	}

	printResponse(resp)
}

func streaming(provider genai.Provider) {
	msg := genai.NewTextMessage(genai.RoleUser, "Count from 1 to 5, one number per line.")
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(64).
		SetTemperature(0.0).
		Build()

	ctx := context.Background()
	respCh, errCh := provider.GenerateStream(ctx, defaultModel, msg, opts)

	fmt.Print("Response: ")
	for resp := range respCh {
		for _, c := range resp.Candidates {
			for _, p := range c.Message.Parts {
				if p.Text != nil {
					fmt.Print(p.Text.Text)
				}
			}
		}
	}
	fmt.Println()

	if err := <-errCh; err != nil {
		log.Printf("Streaming error: %v\n", err)
	}
}

func topKSampling(provider genai.Provider) {
	// top_k is supported by Claude but not by OpenAI.
	// It limits sampling to the top K most likely tokens.
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(128).
		SetTemperature(0.8).
		SetTopK(40).
		Build()

	msg := genai.NewTextMessage(genai.RoleUser, "Give me a creative one-sentence story about a cat.")
	resp, err := provider.Generate(context.Background(), defaultModel, msg, opts)
	if err != nil {
		log.Printf("Generation error: %v\n", err)
		return
	}

	printResponse(resp)
}

func printResponse(resp *genai.GenResponse) {
	for _, candidate := range resp.Candidates {
		fmt.Printf("Candidate %d (finish: %s):\n", candidate.Index, candidate.FinishReason)
		for _, part := range candidate.Message.Parts {
			if part.Text != nil {
				fmt.Println(part.Text.Text)
			}
			if part.FuncCall != nil {
				fmt.Printf("[Tool call] %s(%v)\n", part.FuncCall.FunctionName, part.FuncCall.Arguments)
			}
		}
	}
}
