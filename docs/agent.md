# Agent

The `Agent` represents an AI agent with configuration, tools, and capabilities.

## Structure

```go
type Agent struct {
	Name                     string                 // Name of the agent
	LLM                      *LLM                   // Language model configuration
	SystemPrompts            []SystemPrompt         // System prompts that define the agent's behavior
	StructuredResponseSchema Schema                 // Schema for structured responses
	Tools                    []Tool                 // Available tools the agent can use
	toolImplementations      map[string]ToolImplementation // Tool implementations
	Messages                 []Message              // Current conversation messages
	conversationHistory      []Message              // Complete conversation history
	MaxRetry                 int                    // Maximum number of retry attempts
	Context                  map[string]interface{} // Context information
	MetaData                 map[string]interface{} // Metadata about the agent
	PromptManager            *PromptManager         // Manager for prompt templates
}
```

## Creation

```go
// Create a new agent
agent := sapiens.NewAgent(
    "MyAssistant",       // Name
    llmImplementation,   // LLM implementation
    apiKey,              // API key
    "gemini-2.0-flash",  // Model name
    "google"             // Provider name
)
```

## Key Methods

### System Prompts

```go
// Add a simple system prompt
agent.AddSystemPrompt("You are a helpful assistant that provides concise answers.", "1.0")

// Add a dynamic prompt from a template and data
promptTemplate := `You are an assistant specialized in {{field}} who answers in {{style}} style.`
cardData := map[string]interface{}{
    "field": "machine learning",
    "style": "clear and educational",
}
agent.AddDynamicPrompt(promptTemplate, cardData, "1.1")

// Create and manage reusable prompt templates
template := sapiens.PromptTemplate{
    Name: "expert_assistant",
    Template: `You are a {{level}} {{field}} expert. Help users with {{task}} questions.`,
    Description: "Template for expert assistants in various fields",
    Version: "1.0",
}
agent.AddPromptTemplate(template)

// Create a card from the template
card := sapiens.NewCard("expert_assistant", map[string]interface{}{
    "level": "senior",
    "field": "software engineering", 
    "task": "code review",
})

// Apply the card template as a system prompt
agent.AddDynamicPromptWithCard(card, "1.0")
```

### Managing Tools

```go
// Add tools to the agent
weatherTool := sapiens.Tool{
    Name: "get_weather",
    Description: "Get current weather for a location",
    InputSchema: &sapiens.Schema{...},
}
agent.AddTools(weatherTool)

// Register tool implementation
agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
    location := params["location"].(string)
    // Implementation logic
    return weatherData, nil
})
```

### Context Management

```go
// Set context as a map
agent.SetContext(map[string]interface{}{
    "user_id": "user123",
    "preferences": map[string]string{
        "language": "English",
        "units": "metric",
    },
})

// Set context as a string
agent.SetStringContext(`
User: John Smith
Subscription: Premium
Preferences: Dark mode, Technical responses
`)

// Update existing context
agent.UpdateStringContext(`
User: John Smith
Subscription: Premium (expires in 7 days)
Preferences: Dark mode, Technical responses, Code examples
`)
```

### Structured Output

```go
// Define output schema
schema := sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "summary": {Type: "string", Description: "Brief summary"},
        "details": {Type: "string", Description: "Detailed explanation"},
        "relevance": {Type: "number", Description: "Relevance score 0-10"},
    },
    Required: []string{"summary", "relevance"},
}

// Set schema for structured responses
agent.SetStructuredResponseSchema(schema)
```

### Running the Agent

```go
// Execute a query
ctx := context.Background()
response, err := agent.Run(ctx, "How do I implement a binary search tree in Go?")
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Access the response
fmt.Println("Content:", response.Content)

// Access tool calls if any were made
for _, toolCall := range response.ToolCalls {
    fmt.Printf("Tool called: %s with inputs: %v\n", toolCall.Name, toolCall.InputMap)
}

// Access structured data if a schema was defined
if response.Structured != nil {
    data := response.Structured.(map[string]interface{})
    fmt.Printf("Summary: %s (Relevance: %.1f)\n", data["summary"], data["relevance"])
}

// Access conversation history
history := agent.GetHistory()
fmt.Printf("Conversation has %d turns\n", len(history)/2)
```

## Related Types

- [LLM](llm.md)
- [SystemPrompt](prompt.md)
- [Tool](tool.md)
- [Schema](schema.md)
- [Message](message.md)
- [PromptManager](prompt.md#promptmanager)
