// Package main demonstrates the Ollama GenAI provider for local model inference.
//
// Prerequisites:
//
//	Start Ollama: ollama serve
//	Pull a model: ollama pull llama3
//	go run main.go
//
// Optional environment variables:
//
//	OLLAMA_MODEL — model to use (default: "llama3")
//	OLLAMA_URL   — Ollama server URL (default: "http://localhost:11434/v1")
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
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")

	// Create the Ollama provider
	var provider genai.Provider
	if ollamaURL != "" {
		// Custom URL
		provider = impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
			BaseURL: ollamaURL,
			Models:  []string{model},
		}, nil)
	} else {
		// Default: http://localhost:11434/v1, no auth
		provider = impl.NewOllamaProvider(nil)
	}
	defer provider.Close()

	fmt.Printf("Provider: %s (v%s)\n", provider.Name(), provider.Version())
	fmt.Printf("Description: %s\n", provider.Description())
	fmt.Printf("Model: %s\n\n", model)

	// --- Basic text generation ---
	fmt.Println("=== Basic Generation ===")
	basicGeneration(provider, model)

	// --- Generation with system instructions ---
	fmt.Println("\n=== System Instructions ===")
	systemInstructions(provider, model)

	// --- Streaming generation ---
	fmt.Println("\n=== Streaming ===")
	streaming(provider, model)

	// --- Using prompt templates ---
	fmt.Println("\n=== Prompt Templates ===")
	promptTemplate(provider, model)
}

func basicGeneration(provider genai.Provider, model string) {
	msg := genai.NewTextMessage(genai.RoleUser, "What are the three primary colors? Answer briefly.")
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(128).
		SetTemperature(0.3).
		Build()

	resp, err := provider.Generate(context.Background(), model, msg, opts)
	if err != nil {
		log.Printf("Generation error (is Ollama running? have you pulled %s?): %v\n", model, err)
		return
	}

	printResponse(resp)
	fmt.Printf("Tokens: input=%d, output=%d, total=%d\n",
		resp.Meta.InputTokens, resp.Meta.OutputTokens, resp.Meta.TotalTokens)
}

func systemInstructions(provider genai.Provider, model string) {
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetTemperature(0.5).
		Build()
	opts.Set(genai.OptionSystemInstructions,
		"You are a friendly Linux sysadmin. Respond with brief shell commands and explanations.")

	msg := genai.NewTextMessage(genai.RoleUser, "How do I find the largest files on my disk?")
	resp, err := provider.Generate(context.Background(), model, msg, opts)
	if err != nil {
		log.Printf("Generation error: %v\n", err)
		return
	}

	printResponse(resp)
}

func streaming(provider genai.Provider, model string) {
	msg := genai.NewTextMessage(genai.RoleUser, "Count from 1 to 5, one number per line.")
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(64).
		SetTemperature(0.0).
		Build()

	ctx := context.Background()
	respCh, errCh := provider.GenerateStream(ctx, model, msg, opts)

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

func promptTemplate(provider genai.Provider, model string) {
	// Build a prompt from a template
	store := genai.NewInMemoryPromptStore()
	pt, _ := genai.NewPromptTemplate(
		"explain",
		"explanation",
		"Explain {{.topic}} in {{.style}} style in 2-3 sentences.",
	)
	store.Add(pt)

	msg, err := genai.NewMsgFromPromptId(genai.RoleUser, store, "explain",
		map[string]any{
			"topic": "goroutines",
			"style": "simple",
		})
	if err != nil {
		log.Printf("Prompt error: %v\n", err)
		return
	}

	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetTemperature(0.5).
		Build()

	resp, err := provider.Generate(context.Background(), model, msg, opts)
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
