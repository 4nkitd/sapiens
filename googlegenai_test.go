package sapiens

import (
	"context"
	"os"
	"testing"
)

func TestGoogleGenAI(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GOOGLE_API_KEY environment variable not set")
	}

	// Create a new GoogleGenAI instance
	llm := NewGoogleGenAI(apiKey, "gemini-2.0-flash")

	// Initialize the LLM
	err := llm.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Test basic completion
	ctx := context.Background()
	prompt := "What is the capital of France?"
	response, err := llm.Complete(ctx, prompt)
	t.Logf("Response: %s", response)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if response == "" {
		t.Fatal("Expected non-empty response but got empty string")
	}

	// Test completion with options
	options := map[string]interface{}{
		"temperature": float32(0.2),
		"max_tokens":  500,
	}

	responseWithOpts, err := llm.CompleteWithOptions(ctx, prompt, options)

	if err != nil {
		t.Fatalf("CompleteWithOptions failed: %v", err)
	}

	if responseWithOpts == "" {
		t.Fatal("Expected non-empty response but got empty string")
	}
}
