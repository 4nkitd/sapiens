# Sapiens: Go AI Agent Framework

Sapiens is a lightweight, flexible framework for building AI agents in Go. It provides seamless integration with various LLM providers, including Google Gemini AI and OpenAI, with tools for creating, managing, and deploying intelligent agents.

## Features

- **Multiple LLM Support**: Built-in support for Google Gemini and OpenAI (easily extensible to other providers)
- **Tool Calling**: Define and use tools with input/output schemas for function calling 
- **Structured Output**: Generate structured responses based on defined schemas
- **Conversation Memory**: Built-in conversation history management
- **Dynamic Prompting**: Template-based prompt generation with versioning
- **Context Management**: Add and update contextual information for your agents
- **Streaming Responses**: Support for streaming AI responses in real-time

## Installation

```bash
go get github.com/4nkitd/sapiens
```

## Quick Start

### Basic Agent

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
	// Initialize with Google Gemini (or use other providers)
	apiKey := os.Getenv("GEMINI_API_KEY")
	llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}
	
	// Create a new agent
	agent := sapiens.NewAgent("Assistant", llm, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful assistant that provides concise answers.", "1.0")
	
	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Println(response.Content)
}
```

### Agent with Tools

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
	apiKey := os.Getenv("GEMINI_API_KEY")
	llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}
	
	// Create an agent
	agent := sapiens.NewAgent("WeatherAssistant", llm, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful weather assistant.", "1.0")
	
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
		return map[string]interface{}{
			"temperature": 72,
			"condition": "Sunny",
			"location": location,
		}, nil
	})
	
	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "What's the weather like in Paris?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Println(response.Content)
}
```

### Dynamic Prompting

```go
promptTemplate := `
You are a {{tone}} assistant specialized in {{field}}.
When discussing {{topic}}, keep these points in mind:
{{#points}}
- {{.}}
{{/points}}
`

cardData := map[string]interface{}{
	"tone": "friendly",
	"field": "programming",
	"topic": "Go development",
	"points": []string{
		"Explain concepts clearly",
		"Provide working code examples",
		"Follow Go best practices",
	},
}

dynamicPrompt, err := sapiens.ApplyTemplate(promptTemplate, cardData)
agent.AddSystemPrompt(dynamicPrompt, "1.0")
```

## Documentation

For more detailed documentation, see the [docs](./docs) directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
