// Package main demonstrates the OpenAI GenAI provider.
//
// Prerequisites:
//
//	export OPENAI_API_KEY="sk-..."
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

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create the OpenAI provider (uses Bearer token auth automatically)
	provider := impl.NewOpenAIProvider(apiKey, nil)
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

	// --- JSON output mode ---
	fmt.Println("\n=== JSON Output ===")
	jsonOutput(provider)
}

func basicGeneration(provider genai.Provider) {
	msg := genai.NewTextMessage(genai.RoleUser, "What are the three primary colors? Answer briefly.")
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(128).
		SetTemperature(0.3).
		Build()

	resp, err := provider.Generate(context.Background(), "gpt-4o-mini", msg, opts)
	if err != nil {
		log.Printf("Generation error: %v\n", err)
		return
	}

	printResponse(resp)
	fmt.Printf("Tokens: input=%d, output=%d, total=%d\n",
		resp.Meta.InputTokens, resp.Meta.OutputTokens, resp.Meta.TotalTokens)
}

func systemInstructions(provider genai.Provider) {
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetTemperature(0.5).
		Build()
	opts.Set(genai.OptionSystemInstructions,
		"You are a pirate captain. Respond in pirate speak. Keep answers short.")

	msg := genai.NewTextMessage(genai.RoleUser, "What is the weather like today?")
	resp, err := provider.Generate(context.Background(), "gpt-4o-mini", msg, opts)
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
	respCh, errCh := provider.GenerateStream(ctx, "gpt-4o-mini", msg, opts)

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

func jsonOutput(provider genai.Provider) {
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetOutputMime("application/json").
		Build()
	opts.Set(genai.OptionSystemInstructions,
		"Respond with a JSON object. Include 'country', 'capital', and 'population' fields.")

	msg := genai.NewTextMessage(genai.RoleUser, "Tell me about Japan.")
	resp, err := provider.Generate(context.Background(), "gpt-4o-mini", msg, opts)
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
