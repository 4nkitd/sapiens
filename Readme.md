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
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or sapiens.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider
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

## Documentation

For more detailed documentation, see the [docs](./docs) directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
