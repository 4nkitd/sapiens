package sapiens

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestAgent(t *testing.T) {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()
	response, err := agent.Run(ctx, "What is the capital of France ?")
	if err != nil {
		t.Fatalf("Agent.Run failed: %v", err)
	}

	t.Logf("Response: %s", response.Content)

	if response.Content == "" {
		t.Fatal("Expected non-empty response but got empty string")
	}
}

func TestAgentWithToolCalling(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()

	// Test with tools
	weatherTool := Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &Schema{
			Type: "object",
			Properties: map[string]Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit (celsius or fahrenheit)",
					Enum:        []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"location"},
		},
	}

	agent.AddTools(weatherTool)

	toolResponse, err := agent.Run(ctx, "What's the weather like in Paris?")
	if err != nil {
		t.Fatalf("Agent.Run with tools failed: %v", err)
	}

	t.Logf("Tool calls: %v", toolResponse.ToolCalls)
	t.Logf("Content: %s", toolResponse.Content)

}

func TestAgentWithStructuredResponse(t *testing.T) {
	apiKey :=
		os.Getenv("GEMINI_API_KEY")

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()

	// Test with structured output
	agent = NewAgent("StructuredAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate responses in structured format.", "1.0")

	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"failure_reason": {
				Type:        "string",
				Description: "Reason for failure",
			},
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

	agent.SetStructuredResponseSchema(schema)

	structuredResponse, err := agent.Run(ctx, "what is the population india")
	fmt.Println(err)
	if err != nil {
		t.Fatalf("Agent.Run with structured output failed: %v", err)
	}

	t.Logf("Structured: %v", structuredResponse.Structured)
	t.Logf("Content: %s", structuredResponse.Content)

	if structuredResponse.Structured == nil {
		t.Fatal("Expected structured data but got nil")
	}

}

func TestAgentWithToolImplementation(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant.", "1.0")

	// Test with tools and implementation
	weatherTool := Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &Schema{
			Type: "object",
			Properties: map[string]Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit (celsius or fahrenheit)",
					Enum:        []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"location"},
		},
	}

	agent.AddTools(weatherTool)

	// Register tool implementation
	agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, ok := params["location"].(string)
		if !ok {
			return nil, fmt.Errorf("location must be a string")
		}

		unit := "celsius"
		if unitParam, ok := params["unit"].(string); ok {
			unit = unitParam
		}

		// Mock weather data
		weatherInfo := map[string]interface{}{
			"location":    location,
			"temperature": 22,
			"unit":        unit,
			"condition":   "Rainy",
		}

		fmt.Printf("Weather info: %v\n", weatherInfo)

		return weatherInfo, nil
	})

	ctx := context.Background()
	response, err := agent.Run(ctx, "What's the weather like in Paris?")
	if err != nil {
		t.Fatalf("Agent.Run with tool implementation failed: %v", err)
	}

	t.Logf("Final response: %s", response.Content)

	// Test a follow-up question to see if the agent can use the context
	followUpResponse, err := agent.Run(ctx, "And what should I wear in that weather?")
	if err != nil {
		t.Fatalf("Follow-up question failed: %v", err)
	}

	t.Logf("Follow-up response: %s", followUpResponse.Content)
}
