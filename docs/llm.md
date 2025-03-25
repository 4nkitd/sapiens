# LLM Integration

The `LLMInterface` and related types define how Sapiens interacts with language model providers like Google Gemini and OpenAI.

## Interfaces

### LLMInterface

The `LLMInterface` defines the contract for language model implementations:

```go
type LLMInterface interface {
	// Initialize sets up the language model with provided options
	Initialize() error

	// Set the system prompt for the model
	SetSystemPrompt(prompt SystemPrompt)

	// Generate a response based on a request
	Generate(ctx context.Context, request *Request) (*Response, error)

	// Generate a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)

	// Generate a completion with specific parameters
	CompleteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error)

	// Generate a response based on chat messages
	ChatCompletion(ctx context.Context, messages []Message) (Response, error)

	// Generate a response with tools support
	ChatCompletionWithTools(ctx context.Context, messages []Message, tools []Tool, options map[string]interface{}) (Response, error)

	// Generate a response with tools support and execute tool handlers
	ChatCompletionWithToolsAndHandlers(ctx context.Context, messages []Message, tools []Tool,
		toolHandlers map[string]ToolHandler, options map[string]interface{}) (Response, error)

	// Generate a structured response based on a schema
	StructuredOutput(ctx context.Context, messages []Message, schema Schema) (Response, error)
}
```

## Implementations

### GoogleGenAI

The `GoogleGenAI` implementation provides integration with Google's Generative AI models:

```go
// Create a new GoogleGenAI instance
llm := sapiens.NewGoogleGenAI(apiKey, "gemini-2.0-flash")

// Initialize the client
err := llm.Initialize()
if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}

// Set temperature and max tokens
llm.Temperature = 0.2
llm.MaxTokens = 500
```

## Request and Response Types

### Request

```go
type Request struct {
	Messages                 []Message      // Conversation messages
	Tools                    []Tool         // Available tools
	StructuredResponseSchema *Schema        // Schema for structured response
	SystemPrompts            []SystemPrompt // System prompts
}
```

### Response

```go
type Response struct {
	Content     string      // Text response content
	ToolCalls   []ToolCall  // Tool calls made by the model
	Structured  interface{} // Structured data based on schema
	ToolResults []Message   // Results from tool execution
	Raw         interface{} // Raw response from the LLM
}
```

### Messages

```go
type Message struct {
	Role       string                 // system, user, assistant, or function
	Content    string                 // Message content
	Name       string                 // Optional name
	ToolCallID string                 // ID of the tool call this message is responding to
	ToolCalls  []ToolCall             // Tool calls made by the assistant
	Options    map[string]interface{} // Additional options
}
```

## Tool Integration

Tools allow LLMs to execute functions and use their results:

```go
// Define a tool handler
type ToolHandler func(toolName string, arguments map[string]interface{}) (string, error)

// Define a tool
weatherTool := sapiens.Tool{
    Name:        "get_weather",
    Description: "Get the current weather for a location",
    InputSchema: &sapiens.Schema{
        Type: "object",
        Properties: map[string]sapiens.Schema{
            "location": {Type: "string", Description: "The city name"},
        },
        Required: []string{"location"},
    },
}

// Create a tool handler
toolHandler := func(toolName string, args map[string]interface{}) (string, error) {
    if toolName != "get_weather" {
        return "", fmt.Errorf("unknown tool: %s", toolName)
    }
    
    location := args["location"].(string)
    return fmt.Sprintf("The weather in %s is sunny and 22°C", location), nil
}

// Use the tool with the LLM
handlers := map[string]sapiens.ToolHandler{"get_weather": toolHandler}
response, err := llm.ChatCompletionWithToolsAndHandlers(
    ctx, 
    messages, 
    []sapiens.Tool{weatherTool}, 
    handlers, 
    nil,
)
```

## Structured Output

Generate responses in a specific structured format:

```go
// Define a schema for structured output
schema := sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "title": {Type: "string"},
        "summary": {Type: "string"},
        "keywords": {
            Type: "array",
            Items: &sapiens.Schema{Type: "string"},
        },
    },
    Required: []string{"title", "summary"},
}

// Get a structured response
response, err := llm.StructuredOutput(ctx, messages, schema)
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Access the structured data
data := response.Structured.(map[string]interface{})
fmt.Printf("Title: %s\n", data["title"])
fmt.Printf("Summary: %s\n", data["summary"])
```

## Related Types

- [Agent](agent.md)
- [Tool](tool.md)
- [Schema](schema.md)
- [Message](message.md)
