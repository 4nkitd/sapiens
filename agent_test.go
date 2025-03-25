package sapiens

import (
	"context"
	"fmt"
	"os"
	"strings"
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

func TestAgentMemory(t *testing.T) {
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
	agent := NewAgent("MemoryAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant who remembers previous parts of the conversation.", "1.0")

	ctx := context.Background()

	// Initial message with information to remember
	initialResponse, err := agent.Run(ctx, "My name is Alice and I have a dog named Max.")
	if err != nil {
		t.Fatalf("Initial message failed: %v", err)
	}
	t.Logf("Initial response: %s", initialResponse.Content)

	// First follow-up question that requires memory
	followUpResponse1, err := agent.Run(ctx, "What's my name?")
	if err != nil {
		t.Fatalf("First follow-up question failed: %v", err)
	}
	t.Logf("First follow-up response: %s", followUpResponse1.Content)

	// Check if the response contains the name
	if !containsSubstring(followUpResponse1.Content, "Alice") {
		t.Errorf("Expected response to contain 'Alice', but got: %s", followUpResponse1.Content)
	}

	// Second follow-up to test deeper memory
	followUpResponse2, err := agent.Run(ctx, "What is my dog's name?")
	if err != nil {
		t.Fatalf("Second follow-up question failed: %v", err)
	}
	t.Logf("Second follow-up response: %s", followUpResponse2.Content)

	// Check if the response contains the dog's name
	if !containsSubstring(followUpResponse2.Content, "Max") {
		t.Errorf("Expected response to contain 'Max', but got: %s", followUpResponse2.Content)
	}

	// Test memory persistence with additional context
	furtherResponse, err := agent.Run(ctx, "I also have a cat named Luna. Can you remember both my pets?")
	if err != nil {
		t.Fatalf("Further context message failed: %v", err)
	}
	t.Logf("Response with additional context: %s", furtherResponse.Content)

	// Final memory check
	finalResponse, err := agent.Run(ctx, "What are the names of my pets?")
	if err != nil {
		t.Fatalf("Final memory check failed: %v", err)
	}
	t.Logf("Final memory check response: %s", finalResponse.Content)

	// Check if both pet names are mentioned
	if !containsSubstring(finalResponse.Content, "Max") || !containsSubstring(finalResponse.Content, "Luna") {
		t.Errorf("Expected response to contain both 'Max' and 'Luna', but got: %s", finalResponse.Content)
	}
}

// Helper function to check if a string contains a substring (case insensitive)
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestAgentWithInitialContext(t *testing.T) {
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
	agent := NewAgent("ContextAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that answers questions based on the provided context.", "1.0")

	ctx := context.Background()

	// Prepare initial context about a fictional company
	initialContext := `
		Company: TechNova Solutions
		Founded: 2015
		CEO: Sarah Johnson
		Headquarters: San Francisco, CA
		Employees: 450
		Main Products: 
		- CloudSync (cloud storage solution)
		- DataGuard (cybersecurity platform)
		- SmartAnalytics (business intelligence tool)
		Recent News: TechNova Solutions just raised $50 million in Series C funding led by Acme Ventures.
	`

	// Set the initial context for the agent as a string
	agent.SetStringContext(initialContext)

	// Ask questions that should be answered based on the context
	questions := []string{
		"Who is the CEO of the company?",
		"What products does the company offer?",
		"Where is the company headquartered?",
		"What was the recent funding news?",
	}

	expectedAnswers := []string{
		"Sarah Johnson",
		"CloudSync", "DataGuard", "SmartAnalytics",
		"San Francisco",
		"$50 million", "Series C", "Acme Ventures",
	}

	for i, question := range questions {
		response, err := agent.Run(ctx, question)
		if err != nil {
			t.Fatalf("Question #%d failed: %v", i+1, err)
		}

		t.Logf("Question: %s\nResponse: %s", question, response.Content)

		// Check if the response contains expected information
		for _, expectedText := range expectedAnswers {
			if strings.Contains(question, strings.Split(expectedText, " ")[0]) && !containsSubstring(response.Content, expectedText) {
				t.Errorf("Question #%d: Expected response to contain '%s', but got: %s", i+1, expectedText, response.Content)
			}
		}
	}

	// Test asking about something not in the context
	unrelatedResponse, err := agent.Run(ctx, "What is the capital of France?")
	if err != nil {
		t.Fatalf("Unrelated question failed: %v", err)
	}

	t.Logf("Unrelated question response: %s", unrelatedResponse.Content)

	// Test context update with a string
	updateContext := `
		Updated Information:
		CEO: Michael Chen (promoted from CTO in 2023)
		Employees: Now 520 after recent expansion
		New Product: EdgeCompute (edge computing platform) launched last month
	`

	agent.UpdateStringContext(updateContext)

	// Test with updated context
	updatedResponse, err := agent.Run(ctx, "Who is the current CEO and when did they take the position?")
	if err != nil {
		t.Fatalf("Question after context update failed: %v", err)
	}

	t.Logf("Updated context response: %s", updatedResponse.Content)

	if !containsSubstring(updatedResponse.Content, "Michael Chen") || !containsSubstring(updatedResponse.Content, "2023") {
		t.Errorf("Expected response to mention 'Michael Chen' and '2023', but got: %s", updatedResponse.Content)
	}
}
