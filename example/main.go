package main

import (
	"context"
	"fmt"
	"log"
	"os"

	sapiens "github.com/4nkitd/sapiens"
)

func main() {

	// Initialize GoogleGenAI
	googleAPIKey := os.Getenv("GEMINI_API_KEY")
	googleLLM := sapiens.NewGoogleGenAI(googleAPIKey, "gemini-2.0-flash")
	err := googleLLM.Initialize()
	if err != nil {
		log.Fatalf("Error initializing GoogleGenAI: %v", err)
	}

	// Initialize OpenAI
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	openaiLLM := sapiens.NewOpenAI(openaiAPIKey, "gpt-4o")
	err = openaiLLM.Initialize()
	if err != nil {
		log.Fatalf("Error initializing OpenAI: %v", err)
	}

	// Create an agent with GoogleGenAI
	googleAgent := sapiens.NewAgent("GoogleAgent", googleLLM, googleAPIKey, "gemini-2.0-flash", "google")
	googleAgent.AddSystemPrompt("You are a helpful AI assistant using Google's Gemini.", "1.0")

	// Create an agent with OpenAI
	openAIAgent := sapiens.NewAgent("OpenAIAgent", openaiLLM, openaiAPIKey, "gpt-4o", "openai")
	openAIAgent.AddSystemPrompt("You are a helpful AI assistant using OpenAI's GPT-4.", "1.0")

	// Run the agents
	ctx := context.Background()

	googleResponse, err := googleAgent.Run(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatalf("Error running GoogleAgent: %v", err)
	}
	fmt.Printf("GoogleAgent Response: %s\n", googleResponse.Content)

	openaiResponse, err := openAIAgent.Run(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatalf("Error running OpenAIAgent: %v", err)
	}
	fmt.Printf("OpenAIAgent Response: %s\n", openaiResponse.Content)
}
