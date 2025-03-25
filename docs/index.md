# Sapiens Documentation

Welcome to the Sapiens documentation. Sapiens is a flexible Go framework for building AI agents with support for multiple LLM providers, tool integration, structured outputs, and conversation memory.

## Core Components

- [Agent](agent.md) - AI agent with configuration and capabilities
- [LLM Integration](llm.md) - Interface for language model providers
- [Tool](tool.md) - Capabilities that agents can use
- [Schema](schema.md) - JSON schema for defining data structures
- [Memory](memory.md) - Storage and retrieval system for agents
- [Prompt Templates](prompt.md) - System prompts with templating and versioning

## Getting Started

### Basic Agent Setup

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY")
	llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}
	
	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful assistant specialized in Go programming.", "1.0")
	
	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "How do I create a simple HTTP server in Go?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Println(response.Content)
}
```

### Adding Tools

Tools allow your agent to perform specific functions:

```go
// Define a weather tool
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

// Add the tool to the agent
agent.AddTools(weatherTool)

// Register tool implementation
agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
	location, _ := params["location"].(string)
	
	// This would typically call a real weather API
	return map[string]interface{}{
		"temperature": 72,
		"condition": "Sunny",
		"location": location,
	}, nil
})
```

### Using Structured Output

Request structured responses from your agent:

```go
// Define a response schema
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

// Set the schema on your agent
agent.SetStructuredResponseSchema(schema)

// Get structured responses
response, err := agent.Run(ctx, "What is the population of London?")
if err != nil {
	log.Fatalf("Error: %v", err)
}

// Access the structured data
structuredData := response.Structured
fmt.Printf("Answer: %s (Confidence: %.2f)\n", 
	structuredData.(map[string]interface{})["answer"],
	structuredData.(map[string]interface{})["confidence"])
```

### Adding Context

Provide your agent with additional context:

```go
// Add context information
agent.SetStringContext(`
Company: TechCorp
Founded: 2010
Products: Cloud services, AI solutions, Data analytics
Headquarters: San Francisco
`)

// The agent will use this context when answering questions
response, err := agent.Run(ctx, "When was our company founded?")
```

For more detailed information, see the documentation for each specific component.
