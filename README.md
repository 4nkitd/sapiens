# Sapiens

A simple and powerful Go library for building AI agents with multi-LLM provider support, tool integration, and structured outputs.

## Features

- **Multi-LLM Provider Support**: OpenAI, Google Gemini, Anthropic Claude, and Ollama
- **Tool/Function Calling**: Add custom tools that agents can use
- **Structured Outputs**: Define JSON schemas for structured responses
- **Conversation History**: Automatic message history management
- **Thread-Safe**: Concurrent operations with proper synchronization
- **Recursion Protection**: Prevents infinite tool call loops

## Installation

```bash
go get github.com/4nkitd/sapiens
```

## Quick Start

### Basic Usage

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

    // Create message
    message := NewMessages()

    // Ask a question
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Hello! How are you today?"),
    ))

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### Adding Tools

Tools allow your agent to perform specific functions:

```go
package sapiens

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a weather assistant",
    )

    // Add a weather tool
    agent.AddTool(
        "get_weather",
        "Get current weather for a location",
        map[string]jsonschema.Definition{
            "location": {
                Type:        jsonschema.String,
                Description: "The city and state, e.g. San Francisco, CA",
            },
            "unit": {
                Type: jsonschema.String,
                Enum: []string{"celsius", "fahrenheit"},
            },
        },
        []string{"location"},
        func(parameters map[string]string) string {
            location := parameters["location"]
            unit := parameters["unit"]

            // Your weather API logic here
            return fmt.Sprintf(`{"temperature":"25", "unit":"%s", "location":"%s"}`, unit, location)
        },
    )

    message := NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What's the weather in London?"),
    ))

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### Structured Outputs

Request structured responses using JSON schemas:

```go
package sapiens

import (
    "context"
    "log"
    "os"
)

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )

    // Define response structure
    type Result struct {
        Steps []struct {
            Explanation string `json:"explanation"`
            Output      string `json:"output"`
        } `json:"steps"`
        FinalAnswer string `json:"final_answer"`
    }
    var result Result

    // Set structured response schema
    agent.SetResponseSchema(
        "analysis_result",
        "Structured analysis with steps and final answer",
        true,
        result,
    )

    message := NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Analyze the benefits of renewable energy"),
    ))

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    // Parse the structured response
    err = agent.ParseResponse(resp, &result)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }

    log.Printf("Structured result: %+v", result)
}
```

## Supported LLM Providers

### OpenAI

```go
llm := NewOpenai(os.Getenv("OPENAI_API_KEY"))
```

Default model: `gpt-4.1-2025-04-14`

### Google Gemini

```go
llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
```

Default model: `gemini-2.0-flash`

### Anthropic Claude

```go
llm := NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"))
```

Default model: `claude-sonet-3.5`

### Ollama

```go
llm := NewOllama(
    "http://localhost:11434/v1/",  // Base URL
    "",                            // Auth token (optional)
    "llama2",                      // Model name
)
```

## API Reference

### Agent

#### `NewAgent(ctx, client, model, systemPrompt) *Agent`

Creates a new agent instance.

**Parameters:**
- `ctx`: Context for operations
- `client`: OpenAI-compatible client
- `model`: Model name to use
- `systemPrompt`: System prompt that defines agent behavior

#### `AddTool(name, description, parameters, required, callback) error`

Adds a tool that the agent can use.

**Parameters:**
- `name`: Tool name
- `description`: Tool description
- `parameters`: JSON schema definition for parameters
- `required`: Required parameter names
- `callback`: Function to execute when tool is called

#### `SetResponseSchema(name, description, strict, schema) *ResponseFormat`

Sets up structured output schema.

**Parameters:**
- `name`: Schema name
- `description`: Schema description
- `strict`: Whether to enforce strict validation
- `schema`: Go struct defining the response structure

#### `Ask(messages) (ChatCompletionResponse, error)`

Sends messages to the agent and returns response.

#### `ParseResponse(response, target) error`

Parses a structured response into a Go struct.

### Messages

#### `NewMessages() *Messages`

Creates a new message builder.

#### `UserMessage(content) ChatCompletionMessage`

Creates a user message.

#### `AgentMessage(content) ChatCompletionMessage`

Creates an assistant message.

#### `ToolMessage(id, name, content) ChatCompletionMessage`

Creates a tool response message.

#### `MergeMessages(...messages) []ChatCompletionMessage`

Combines multiple messages into a slice.

## Advanced Features

### Multiple Tools

You can add multiple tools to an agent:

```go
// Add weather tool
agent.AddTool("get_weather", "Get weather info", weatherParams, []string{"location"}, weatherFunc)

// Add currency tool
agent.AddTool("convert_currency", "Convert currency", currencyParams, []string{"amount", "from", "to"}, currencyFunc)

// The agent will automatically choose which tools to use
```

### Tool Call Recursion Protection

The agent automatically prevents infinite tool call loops with a maximum depth of 5 calls.

### Thread Safety

All agent operations are thread-safe and can be used concurrently.

## Examples

See the test files for more examples:
- `agent_ask_test.go` - Basic agent usage with tools
- `agent_multiple_tools_test.go` - Multiple tool integration
- `agent_structured_resp_test.go` - Structured response handling

## Environment Variables

Set the appropriate API key for your chosen provider:

```bash
export OPENAI_API_KEY="your-openai-key"
export GEMINI_API_KEY="your-gemini-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
```

## Dependencies

- `github.com/sashabaranov/go-openai` - OpenAI Go client (used for all providers)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
