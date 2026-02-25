package main

import (
	"fmt"

	"oss.nandlabs.io/golly/genai"
)

func main() {
	// --- Creating Messages ---
	fmt.Println("=== Creating GenAI Messages ===")

	// Simple text message from a prompt string
	userMsg := genai.NewMsgFromPrompt(genai.RoleUser, "greeting", "Hello! Tell me about Go programming.")
	fmt.Printf("Message: role=%s, parts=%d\n", userMsg.Role, len(userMsg.Parts))
	for _, part := range userMsg.Parts {
		if part.Text != nil {
			fmt.Printf("  Part: name=%s, text=%s\n", part.Name, part.Text.Text)
		}
	}

	// System message
	systemMsg := genai.NewMsgFromPrompt(genai.RoleSystem, "system", "You are a helpful Go programming assistant.")
	fmt.Printf("System: role=%s\n", systemMsg.Role)

	// Message with binary data (e.g., image)
	imgData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG header bytes
	binMsg := genai.NewBinMessage(genai.RoleUser, "photo", imgData, "image/jpeg")
	fmt.Printf("Binary message: role=%s, parts=%d\n", binMsg.Role, len(binMsg.Parts))

	// Message with file reference
	fileMsg := genai.NewFileMessage(genai.RoleUser, "document", "gs://bucket/doc.pdf", "application/pdf")
	fmt.Printf("File message: role=%s, uri=%s\n", fileMsg.Role, fileMsg.Parts[0].File.URI)

	// JSON message
	data := map[string]interface{}{
		"name":  "Alice",
		"items": []string{"Go", "Rust", "Python"},
	}
	jsonMsg, err := genai.NewJsonMessage(genai.RoleUser, "structured-data", data)
	if err != nil {
		fmt.Println("JSON message error:", err)
	} else {
		fmt.Printf("JSON message: role=%s, parts=%d\n", jsonMsg.Role, len(jsonMsg.Parts))
	}

	// --- Adding Parts to Existing Messages ---
	fmt.Println("\n=== Adding Parts to Messages ===")
	msg := genai.NewMsgFromPrompt(genai.RoleUser, "multi", "Analyze this image:")
	genai.AddBinPart(msg, "image", imgData, "image/png")
	genai.AddTextPart(msg, "followup", "What do you see?")
	fmt.Printf("Multi-part message: %d parts\n", len(msg.Parts))
	for i, part := range msg.Parts {
		fmt.Printf("  Part %d: name=%s\n", i, part.Name)
	}

	// --- Prompt Store ---
	fmt.Println("\n=== In-Memory Prompt Store ===")
	store := genai.NewInMemoryPromptStore()
	pt, _ := genai.NewPromptTemplate("welcome", "welcome", "Welcome {{.name}} to {{.app}}!")
	store.Add(pt)

	welcomeMsg, err := genai.NewMsgFromPromptId(
		genai.RoleAssistant,
		store,
		"welcome",
		map[string]any{"name": "Alice", "app": "Golly"},
	)
	if err != nil {
		fmt.Println("Prompt error:", err)
	} else {
		if welcomeMsg.Parts[0].Text != nil {
			fmt.Printf("From prompt store: %s\n", welcomeMsg.Parts[0].Text.Text)
		}
	}

	// --- Options ---
	fmt.Println("\n=== GenAI Options ===")
	opts := &genai.Options{}
	opts.Set(genai.OptionMaxTokens, 1024)
	opts.Set("temperature", 0.7)
	opts.Set("top_p", 0.9)
	fmt.Printf("Options: maxTokens=%d, temperature=%.1f, topP=%.1f\n",
		opts.GetMaxTokens(0), opts.GetTemperature(0), opts.GetTopP(0))

	// --- Provider interface (for reference) ---
	fmt.Println("\n=== Provider Interface ===")
	fmt.Println("The genai.Provider interface defines:")
	fmt.Println("  Name() string")
	fmt.Println("  Description() string")
	fmt.Println("  Version() string")
	fmt.Println("  Models() []string")
	fmt.Println("  Generate(ctx, model, message, options) (*GenResponse, error)")
	fmt.Println("  GenerateStream(ctx, model, message, options) (<-chan *GenResponse, <-chan error)")
	fmt.Println("  Close() error")
	fmt.Println("\nSee genai/impl for OpenAI and Ollama implementations.")
}
