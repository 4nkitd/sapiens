# Message

The `Message` type represents a single message in a conversation between a user and an AI assistant.

## Structure

```go
type Message struct {
	Role       string                 // system, user, assistant, or function
	Content    string                 // Message content
	Name       string                 // Optional name (used with function role)
	ToolCallID string                 // ID of the tool call this message is responding to
	ToolCalls  []ToolCall             // Tool calls made by the assistant
	Options    map[string]interface{} // Additional model-specific options
}
```

## Usage

```go
// Create a system message
systemMsg := sapiens.Message{
    Role:    "system",
    Content: "You are a helpful assistant specialized in programming.",
}

// Create a user message
userMsg := sapiens.Message{
    Role:    "user",
    Content: "How do I sort an array in JavaScript?",
}

// Create an assistant message
assistantMsg := sapiens.Message{
    Role:    "assistant",
    Content: "To sort an array in JavaScript, you can use the built-in sort() method...",
}

// Create a function/tool response message
toolMsg := sapiens.Message{
    Role:       "function",
    Name:       "get_weather",
    Content:    `{"temperature": 22, "condition": "sunny", "location": "San Francisco"}`,
    ToolCallID: "call_123456",
}

// Assistant message with tool calls
assistantWithToolMsg := sapiens.Message{
    Role:    "assistant",
    Content: "Let me check the weather for you.",
    ToolCalls: []sapiens.ToolCall{
        {
            ID:   "call_123456",
            Name: "get_weather",
            InputMap: map[string]interface{}{
                "location": "San Francisco",
            },
        },
    },
}
```

## Message Roles

- **system**: Provides instructions or context to the assistant about how it should behave
- **user**: Represents input from the user
- **assistant**: Represents responses from the assistant
- **function/tool**: Represents results from a tool or function call

## Conversation Flow

A typical conversation flow with tool usage looks like this:

1. System message sets the behavior of the assistant
2. User sends a message asking for information
3. Assistant responds by making a tool call
4. Tool execution results are added as function/tool messages
5. Assistant provides a final response using the tool results

Example:

```go
messages := []sapiens.Message{
    {
        Role:    "system",
        Content: "You are a helpful assistant with access to weather information.",
    },
    {
        Role:    "user",
        Content: "What's the weather like in Paris?",
    },
    {
        Role:    "assistant",
        Content: "I'll check the weather in Paris for you.",
        ToolCalls: []sapiens.ToolCall{
            {
                ID:   "call_abc123",
                Name: "get_weather",
                InputMap: map[string]interface{}{
                    "location": "Paris",
                },
            },
        },
    },
    {
        Role:       "function",
        Name:       "get_weather",
        ToolCallID: "call_abc123",
        Content:    `{"temperature": 18, "condition": "rainy", "location": "Paris"}`,
    },
    {
        Role:    "assistant",
        Content: "The current weather in Paris is rainy with a temperature of 18°C.",
    },
}
```

## Related Types

- [Agent](agent.md)
- [ToolCall](llm.md#tool-integration)
- [Response](llm.md#response)
