# Sapiens Documentation

Welcome to the Sapiens documentation. Sapiens is a Go library for building AI agents with support for multiple LLM providers, tool integration, and structured outputs.

## Getting Started

Sapiens provides a simple API for creating AI agents that can work with different language model providers through a unified interface.

### Basic Example

```go
package sapiens

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    // Initialize LLM provider
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Create and send message
    message := NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Hello! How can you help me?"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

## Core Components

### Agent
The [Agent](agent.md) is the main interface for interacting with AI models. It manages conversation history, tools, and coordinates with LLM providers.

### LLM Providers
Sapiens supports multiple [LLM providers](llm.md):
- OpenAI (GPT models)
- Google Gemini
- Anthropic Claude
- Ollama (local models)

### Tools
[Tools](tool.md) extend agent capabilities by allowing them to perform specific functions like API calls, calculations, or data retrieval. Includes support for both regular tools and MCP (Model Context Protocol) servers.

### Messages
The [Messages](message.md) system provides utilities for creating and managing conversation messages.

### Structured Outputs
Define [JSON schemas](schema.md) to get structured responses from agents instead of plain text.

## Key Features

### Multi-Provider Support
```go
// OpenAI
llm := NewOpenai(os.Getenv("OPENAI_API_KEY"))

// Google Gemini
llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

// Anthropic
llm := NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"))

// Ollama
llm := NewOllama("http://localhost:11434/v1/", "", "llama2")
```

### MCP (Model Context Protocol) Support
```go
// Connect to MCP server for external tools
err := agent.AddMCP("http://localhost:8080/sse", nil)
if err != nil {
    log.Printf("MCP connection failed: %v", err)
}

// Agent automatically uses both regular and MCP tools
fmt.Printf("Agent has %d regular tools and %d MCP tools\n", 
    len(agent.Tools), len(agent.McpTools))
```

### Tool Integration
```go
agent.AddTool(
    "get_weather",
    "Get current weather for a location",
    map[string]jsonschema.Definition{
        "location": {
            Type:        jsonschema.String,
            Description: "The city and state",
        },
    },
    []string{"location"},
    func(parameters map[string]string) string {
        // Tool implementation
        return `{"temperature": "25°C", "condition": "sunny"}`
    },
)
```

### Structured Responses
```go
type WeatherReport struct {
    Temperature string `json:"temperature"`
    Condition   string `json:"condition"`
    Humidity    string `json:"humidity"`
}

var report WeatherReport
agent.SetResponseSchema("weather_report", "Weather information", true, report)
```

## Architecture

Sapiens uses the OpenAI client library as a common interface for all providers. Each provider implements:

- Client creation with appropriate base URLs and authentication
- Model name defaults
- Provider-specific configurations

The Agent orchestrates:
- Message history management
- Tool execution and recursion protection
- Structured output parsing
- Thread-safe operations

## Quick Reference

| Component | Purpose | Example |
|-----------|---------|---------|
| `NewAgent()` | Create an agent | `NewAgent(ctx, client, model, prompt)` |
| `AddTool()` | Add tool capability | `agent.AddTool(name, desc, params, required, func)` |
| `AddMCP()` | Connect to MCP server | `agent.AddMCP("http://localhost:8080/sse", nil)` |
| `SetResponseSchema()` | Enable structured output | `agent.SetResponseSchema(name, desc, strict, schema)` |
| `Ask()` | Send messages | `agent.Ask(messages)` |
| `NewMessages()` | Message builder | `msg := NewMessages()` |

## Next Steps

1. **[Set up your first agent](agent.md#creation)** - Learn how to create and configure agents
2. **[Add tools](tool.md)** - Extend agent capabilities with custom functions and MCP servers
3. **[Use structured outputs](schema.md)** - Get structured data instead of plain text
4. **[Explore examples](examples.md)** - See real-world usage patterns including MCP integration

## API Documentation

For detailed API information, see:
- [Agent API](agent.md)
- [LLM Providers](llm.md)
- [Tool System](tool.md)
- [Message Handling](message.md)
- [Schema Definitions](schema.md)