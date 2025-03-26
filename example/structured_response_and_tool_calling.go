package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
)

func mains() {

	apiKey := os.Getenv("GEMINI_API_KEY")

	llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	weatherTool := sapiens.Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &sapiens.Schema{
			Type: "object",
			Properties: map[string]sapiens.Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
			},
			Required: []string{"location"},
		},
	}

	schema := sapiens.Schema{
		Type: "object",
		Properties: map[string]sapiens.Schema{
			"answer": {
				Type:        "string",
				Description: "The answer to the user's question",
			},
			"confidence": {
				Type:        "number",
				Description: "Confidence score from 0 to 1",
			},
		},
		Required: []string{"answer", "confidence"},
	}

	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful assistant and a good friend.", "1.0")
	agent.AddTools(weatherTool)
	agent.SetStructuredResponseSchema(schema)
	_ = schema
	_ = weatherTool
	// Register tool implementation
	agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, _ := params["location"].(string)

		// This would typically call a real weather API
		return map[string]interface{}{
			"temperature": 72,
			"condition":   "Sunny",
			"location":    location,
		}, nil
	})

	response, err := agent.Run(context.Background(), "whats the wather like in Delhi ?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Structured ::", response.Structured)
	fmt.Println("Content ::", response.Content)

}
