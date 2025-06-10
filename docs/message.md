# Messages

The Messages system in Sapiens provides utilities for creating and managing conversation messages in the proper format for different LLM providers. It abstracts the complexity of message formatting and provides a simple interface for building conversations.

## Overview

The Messages system handles:
- Creating properly formatted messages for different roles (user, assistant, tool)
- Managing message types and content
- Merging multiple messages into conversation flows
- Ensuring compatibility across different LLM providers

## Messages Structure

```go
type Messages struct {
    // Empty struct - acts as a factory for creating messages
}
```

The Messages struct serves as a factory for creating `openai.ChatCompletionMessage` objects with the correct formatting.

## Creating Messages

### `NewMessages() *Messages`

Creates a new Messages instance for building conversation messages.

```go
message := NewMessages()
```

## Message Types

### User Messages

Create messages from the user/human perspective.

#### `UserMessage(content) ChatCompletionMessage`

```go
func (a *Messages) UserMessage(msg string) openai.ChatCompletionMessage
```

**Example:**
```go
message := NewMessages()
userMsg := message.UserMessage("Hello, how can you help me today?")
```

### Assistant Messages

Create messages from the AI assistant perspective.

#### `AgentMessage(content) ChatCompletionMessage`

```go
func (a *Messages) AgentMessage(msg string) openai.ChatCompletionMessage
```

**Example:**
```go
message := NewMessages()
assistantMsg := message.AgentMessage("I'm here to help! What would you like to know?")
```

### Tool Messages

Create messages representing tool/function call responses.

#### `ToolMessage(id, name, content) ChatCompletionMessage`

```go
func (a *Messages) ToolMessage(id, name, msg string) openai.ChatCompletionMessage
```

**Parameters:**
- `id`: The tool call ID from the LLM's request
- `name`: The name of the tool that was called
- `msg`: The response content from the tool

**Example:**
```go
message := NewMessages()
toolMsg := message.ToolMessage("call_123", "get_weather", `{"temperature":"25°C", "condition":"sunny"}`)
```

## Combining Messages

### `MergeMessages(...messages) []ChatCompletionMessage`

Combines multiple messages into a single slice for sending to the agent.

```go
func (a *Messages) MergeMessages(messages ...openai.ChatCompletionMessage) []openai.ChatCompletionMessage
```

**Example:**
```go
message := NewMessages()

messages := message.MergeMessages(
    message.UserMessage("What's the weather like?"),
    message.AgentMessage("I'll check the weather for you."),
    message.ToolMessage("call_123", "get_weather", `{"temp":"25°C"}`),
)
```

## Usage Patterns

### Simple Conversation

```go
package main

import (
    "context"
    "fmt"
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
    
    // Create message builder
    message := NewMessages()
    
    // Send a simple user message
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Hello! Can you help me with Go programming?"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### Multi-Turn Conversation

```go
func multiTurnConversation() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    message := NewMessages()
    
    // First interaction
    resp1, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What is Go programming language?"),
    ))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // The agent automatically maintains conversation history
    // Second interaction builds on the first
    resp2, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Can you show me a simple example?"),
    ))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("First response:", resp1.Choices[0].Message.Content)
    fmt.Println("Second response:", resp2.Choices[0].Message.Content)
}
```

### Building Complex Conversations

```go
func complexConversation() {
    message := NewMessages()
    
    // Build a conversation with multiple message types
    conversationHistory := message.MergeMessages(
        message.UserMessage("I need help with a programming problem."),
        message.AgentMessage("I'd be happy to help! What programming problem are you working on?"),
        message.UserMessage("I need to create a REST API in Go."),
        message.AgentMessage("Great! Let me help you with that. What specific part do you need help with?"),
        message.UserMessage("How do I handle JSON requests?"),
    )
    
    // Send the entire conversation context
    resp, err := agent.Ask(conversationHistory)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### Working with Tools

When tools are involved, the message flow becomes more complex, but the Messages system handles it seamlessly:

```go
func toolConversation() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant with access to tools",
    )
    
    // Add a tool (simplified)
    agent.AddTool("get_weather", "Get weather info", weatherParams, []string{"location"}, weatherFunc)
    
    message := NewMessages()
    
    // User asks for weather - this will trigger tool usage automatically
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What's the weather in New York?"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // The agent handles:
    // 1. User message
    // 2. Tool call execution
    // 3. Tool response processing
    // 4. Final response generation
    
    fmt.Println("Final response:", resp.Choices[0].Message.Content)
}
```

## Message Format Details

### User Message Format

```go
openai.ChatCompletionMessage{
    Role:    openai.ChatMessageRoleUser,
    Content: "Your message content here",
}
```

### Assistant Message Format

```go
openai.ChatCompletionMessage{
    Role:    openai.ChatMessageRoleAssistant,
    Content: "Assistant response content",
}
```

### Tool Message Format

```go
openai.ChatCompletionMessage{
    Role:       openai.ChatMessageRoleTool,
    Content:    "Tool response content",
    ToolCallID: "call_id_from_llm",
    Name:       "tool_name",
}
```

## Best Practices

### 1. Use the Message Builder

Always use the Messages builder for consistency:

```go
// Good
message := NewMessages()
userMsg := message.UserMessage("Hello")

// Avoid direct creation
// userMsg := openai.ChatCompletionMessage{...}
```

### 2. Merge Messages for Clarity

Use `MergeMessages` to make conversation flows clear:

```go
messages := message.MergeMessages(
    message.UserMessage("Question 1"),
    message.UserMessage("Question 2"),
    message.UserMessage("Question 3"),
)
```

### 3. Let the Agent Handle Tool Messages

Don't manually create tool messages - the agent handles tool call responses automatically:

```go
// The agent automatically handles this flow:
// 1. User message
// 2. LLM decides to call tool
// 3. Tool executes
// 4. Tool response added as tool message
// 5. Final response generated
```

### 4. Build Conversations Incrementally

For complex scenarios, build conversations step by step:

```go
message := NewMessages()
conversation := []openai.ChatCompletionMessage{}

// Add messages incrementally
conversation = append(conversation, message.UserMessage("Start conversation"))
conversation = append(conversation, message.AgentMessage("Agent response"))
conversation = append(conversation, message.UserMessage("Follow-up"))

// Send all at once
resp, err := agent.Ask(conversation)
```

## Error Handling

The Messages system is simple and rarely produces errors, but you should handle agent errors:

```go
message := NewMessages()
resp, err := agent.Ask(message.MergeMessages(
    message.UserMessage("Your question here"),
))

if err != nil {
    switch {
    case strings.Contains(err.Error(), "context"):
        log.Printf("Context error: %v", err)
    case strings.Contains(err.Error(), "tool"):
        log.Printf("Tool execution error: %v", err)
    default:
        log.Printf("General error: %v", err)
    }
    return
}
```

## Complete Example

Here's a complete example showing different message types and patterns:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
    // Setup
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Add a simple tool
    agent.AddTool(
        "calculate",
        "Perform mathematical calculations",
        map[string]jsonschema.Definition{
            "expression": {
                Type:        jsonschema.String,
                Description: "Math expression to evaluate",
            },
        },
        []string{"expression"},
        func(params map[string]string) string {
            // Simple calculator logic
            return `{"result": "42", "expression": "` + params["expression"] + `"}`
        },
    )
    
    // Create message builder
    message := NewMessages()
    
    // Example 1: Simple question
    fmt.Println("=== Simple Question ===")
    resp1, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What is artificial intelligence?"),
    ))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    fmt.Println("Response:", resp1.Choices[0].Message.Content)
    
    // Example 2: Question that requires tool usage
    fmt.Println("\n=== Tool Usage ===")
    resp2, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What is 15 * 37 + 125?"),
    ))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    fmt.Println("Response:", resp2.Choices[0].Message.Content)
    
    // Example 3: Multi-message conversation
    fmt.Println("\n=== Multi-Message ===")
    resp3, err := agent.Ask(message.MergeMessages(
        message.UserMessage("I'm learning Go programming."),
        message.UserMessage("Can you explain goroutines?"),
        message.UserMessage("Show me a simple example."),
    ))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    fmt.Println("Response:", resp3.Choices[0].Message.Content)
}
```

## Related Documentation

- [Agent](agent.md) - Using messages with agents
- [Tools](tool.md) - How tool messages are handled
- [Examples](examples.md) - More message usage patterns