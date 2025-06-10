# Sapiens API Documentation

This document provides detailed information about the Sapiens API.

## Agent

The `Agent` is the core interface for interacting with AI models.

### Creating a New Agent

```go
agent := sapiens.NewAgent(options ...sapiens.AgentOption)
```

#### Options

- `WithModel(model string)`: Specifies which model to use (e.g., "gpt-4", "claude-2")
- `WithAPIKey(apiKey string)`: Sets the API key for the provider
- `WithProvider(provider string)`: Sets the AI provider ("openai" by default)
- `WithMemory()`: Enables conversation memory
- `WithTools(...Tool)`: Adds capability tools to the agent
- `WithTemperature(temp float64)`: Sets the temperature for response generation
- `WithMaxTokens(tokens int)`: Sets the maximum number of tokens in the response
- `WithSystemPrompt(prompt string)`: Sets a custom system prompt

### Methods

#### Respond

```go
response, err := agent.Respond(ctx context.Context, message string) (string, error)
```

Sends a message to the agent and returns its response as a string.

#### RespondStream

```go
stream, err := agent.RespondStream(ctx context.Context, message string) (Stream, error)
```

Sends a message to the agent and returns a stream of response chunks.

#### Stream Interface

```go
type Stream interface {
    Recv() (string, error)
    Close() error
}
```

The `Recv()` method returns the next chunk of the response, or an error when the stream ends.

## Tools

Tools extend the agent's capabilities by allowing it to perform specific functions.

### Creating a Tool

```go
tool := sapiens.NewTool(
    name string,
    description string,
    handler func(ctx context.Context, args map[string]interface{}) (string, error),
)
```

- `name`: A unique identifier for the tool
- `description`: A description of what the tool does
- `handler`: A function that implements the tool's logic

### Tool Parameters

Define the parameters that your tool accepts using a JSON schema:

```go
tool := sapiens.NewTool(
    "weather",
    "Get weather information for a location",
    weatherHandler,
).WithParameters(map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "location": map[string]interface{}{
            "type": "string",
            "description": "City name or zip code",
        },
        "units": map[string]interface{}{
            "type": "string",
            "enum": []string{"metric", "imperial"},
            "default": "metric",
        },
    },
    "required": []string{"location"},
})
```

## Memory

When enabled with `WithMemory()`, the agent maintains conversation history.

### Custom Memory Implementations

You can provide a custom memory implementation:

```go
type Memory interface {
    Add(message Message) error
    GetMessages() ([]Message, error)
    Clear() error
}

agent := sapiens.NewAgent(
    sapiens.WithCustomMemory(myMemoryImplementation),
)
```

## Error Handling

Sapiens provides specific error types:

- `ErrInvalidAPIKey`: Invalid or missing API key
- `ErrInvalidModel`: Specified model is not supported
- `ErrProviderUnavailable`: AI provider service is unavailable
- `ErrContextCanceled`: Operation was canceled via context
- `ErrRateLimited`: Request was rate limited by the provider
- `ErrToolExecution`: Error occurred during tool execution
