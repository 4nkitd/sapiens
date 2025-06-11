# Agent

The `Agent` is the core component of Sapiens that manages AI interactions, conversation history, tool execution, and structured outputs. It provides a unified interface for working with different LLM providers.

## Structure

```go
type Agent struct {
    MessagesHistory          []openai.ChatCompletionMessage
    Context                  context.Context
    Llm                      *openai.Client
    Model                    string
    SystemPrompt             string
    StructuredResponseSchema *openai.ChatCompletionResponseFormat
    Tools                    []AgentTool
    McpClient                *McpClient
    McpTools                 []mcp.Tool
    Request                  openai.ChatCompletionRequest
    mu                       sync.Mutex
    maxToolCallDepth         int
    currentDepth             int
}
```

## Creating an Agent

```go
func NewAgent(ctx context.Context, llm *openai.Client, model string, systemPrompt string) *Agent
```

**Parameters:**
- `ctx`: Context for operations and cancellation
- `llm`: OpenAI-compatible client (from any provider)
- `model`: Model name to use (e.g., "gpt-4", "gemini-2.0-flash")
- `systemPrompt`: System prompt that defines the agent's behavior and personality

**Example:**
```go
package sapiens

import (
    "context"
    "os"
)

func main() {
    // Create LLM provider
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant specialized in technical support",
    )
}
```

## Adding Tools

Tools extend the agent's capabilities by allowing it to call external functions. Sapiens supports both regular tools (local functions) and MCP (Model Context Protocol) tools from external servers.

### `AddTool(name, description, parameters, required, callback) error`

```go
func (a *Agent) AddTool(
    name string,
    description string,
    tool_parameters map[string]jsonschema.Definition,
    required_params []string,
    funx AgentFunc,
) error
```

**Parameters:**
- `name`: Unique identifier for the tool
- `description`: Description of what the tool does
- `tool_parameters`: JSON schema definitions for parameters
- `required_params`: List of required parameter names
- `funx`: Callback function that implements the tool logic

**Example:**
```go
import "github.com/sashabaranov/go-openai/jsonschema"

// Add a weather tool
agent.AddTool(
    "get_weather",
    "Get current weather information for a specific location",
    map[string]jsonschema.Definition{
        "location": {
            Type:        jsonschema.String,
            Description: "The city and state/country, e.g. San Francisco, CA",
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
        if unit == "" {
            unit = "celsius"
        }
        
        // Call your weather API here
        return fmt.Sprintf(`{"temperature":"25", "unit":"%s", "location":"%s", "condition":"sunny"}`, unit, location)
    },
)
```

### Tool Callback Function

```go
type AgentFunc func(parameters map[string]string) string
```

The callback function receives a map of parameter names to values and should return a JSON string with the tool's response.

## Adding MCP Tools

Connect to MCP (Model Context Protocol) servers to use external tools and services.

### `AddMCP(url, headers) error`

```go
func (a *Agent) AddMCP(url string, customHeaders map[string]string) error
```

**Parameters:**
- `url`: MCP server URL (typically an SSE endpoint like `http://localhost:8080/sse`)
- `customHeaders`: Optional custom headers for authentication

**Example:**
```go
// Connect to MCP server
err := agent.AddMCP("http://localhost:8080/sse", nil)
if err != nil {
    log.Printf("Failed to connect to MCP server: %v", err)
    return
}

fmt.Printf("Agent now has %d MCP tools available\n", len(agent.McpTools))
```

**With Authentication:**
```go
headers := map[string]string{
    "Authorization": "Bearer your-jwt-token",
    "X-API-Key":     "your-api-key",
}

err := agent.AddMCP("https://secure-mcp-server.com/sse", headers)
if err != nil {
    log.Printf("Failed to connect to authenticated MCP server: %v", err)
}
```

### MCP Tool Auto-Discovery

When you connect to an MCP server, tools are automatically:
- **Discovered**: Server provides available tools and their schemas
- **Converted**: MCP schemas are converted to OpenAI-compatible format
- **Integrated**: MCP tools work seamlessly alongside regular tools
- **Callable**: Agent can use MCP tools just like regular tools

## Structured Responses

Configure the agent to return structured data instead of plain text.

### `SetResponseSchema(name, description, strict, schema) *ChatCompletionResponseFormat`

```go
func (a *Agent) SetResponseSchema(
    name string,
    description string,
    strict bool,
    defined_schema interface{},
) *openai.ChatCompletionResponseFormat
```

**Parameters:**
- `name`: Name for the response schema
- `description`: Description of the schema purpose
- `strict`: Whether to enforce strict schema validation
- `defined_schema`: Go struct that defines the response structure

**Example:**
```go
// Define response structure
type AnalysisResult struct {
    Steps []struct {
        Explanation string `json:"explanation"`
        Output      string `json:"output"`
    } `json:"steps"`
    FinalAnswer string `json:"final_answer"`
    Confidence  float64 `json:"confidence"`
}

var result AnalysisResult

// Set the schema
agent.SetResponseSchema(
    "analysis_result",
    "Structured analysis with reasoning steps",
    true,
    result,
)
```

### `ParseResponse(response, target) error`

Parse a structured response into a Go struct.

```go
func (a *Agent) ParseResponse(
    agent_response openai.ChatCompletionResponse,
    defined_schema interface{},
) error
```

**Example:**
```go
var result AnalysisResult
err := agent.ParseResponse(resp, &result)
if err != nil {
    log.Fatalf("Parse error: %v", err)
}

fmt.Printf("Final Answer: %s (Confidence: %.2f)\n", result.FinalAnswer, result.Confidence)
```

## Asking Questions

### `Ask(messages) (ChatCompletionResponse, error)`

Send messages to the agent and get a response.

```go
func (a *Agent) Ask(user_messages []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error)
```

**Example:**
```go
// Create message helper
message := NewMessages()

// Send a question
resp, err := agent.Ask(message.MergeMessages(
    message.UserMessage("What's the weather in London and convert 50 USD to EUR?"),
))

if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Println("Response:", resp.Choices[0].Message.Content)
```

## Message Management

Use the Messages helper to create properly formatted messages:

```go
// Create message builder
message := NewMessages()

// Create different types of messages
userMsg := message.UserMessage("Hello, how are you?")
agentMsg := message.AgentMessage("I'm doing well, thank you!")
toolMsg := message.ToolMessage("tool_call_id", "tool_name", "tool_response")

// Combine messages
messages := message.MergeMessages(userMsg, agentMsg)
```

## Advanced Features

### Thread Safety

All agent operations are thread-safe and protected by mutexes. You can safely use the same agent instance across multiple goroutines.

### Tool Call Recursion Protection

The agent automatically prevents infinite tool call loops:
- Maximum depth of 5 tool call levels
- Automatic termination when depth is exceeded
- Error reporting for exceeded recursion

### Conversation History

The agent automatically manages conversation history:
- Stores all messages (user, assistant, tool responses)
- Maintains context across multiple interactions
- History is preserved throughout the agent's lifetime

### Multiple Tools and MCP Integration

You can add multiple tools to a single agent, including both regular tools and MCP tools:

```go
// Add regular tools
agent.AddTool("get_weather", "Get weather info", weatherParams, []string{"location"}, weatherFunc)
agent.AddTool("convert_currency", "Convert currency", currencyParams, []string{"amount", "from", "to"}, currencyFunc)
agent.AddTool("calculate", "Perform calculations", calcParams, []string{"expression"}, calcFunc)

// Connect to MCP servers for additional tools
err := agent.AddMCP("http://payments-server:8080/sse", nil)
if err != nil {
    log.Printf("Payment server not available: %v", err)
}

err = agent.AddMCP("http://analytics-server:9090/sse", map[string]string{
    "Authorization": "Bearer token",
})
if err != nil {
    log.Printf("Analytics server not available: %v", err)
}

// The agent will automatically choose which tools to use (regular or MCP) based on the user's request
fmt.Printf("Agent has %d regular tools and %d MCP tools\n", len(agent.Tools), len(agent.McpTools))
```

## Error Handling

The agent provides detailed error information:

```go
resp, err := agent.Ask(messages)
if err != nil {
    // Handle different types of errors
    switch {
    case strings.Contains(err.Error(), "tool call"):
        log.Printf("Tool execution error: %v", err)
    case strings.Contains(err.Error(), "maximum tool call depth"):
        log.Printf("Recursion limit exceeded: %v", err)
    default:
        log.Printf("General error: %v", err)
    }
    return
}
```

## Complete Example

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
    // Initialize LLM
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant that can check weather and convert currencies",
    )
    
    // Add weather tool
    agent.AddTool(
        "get_weather",
        "Get current weather for a location",
        map[string]jsonschema.Definition{
            "location": {
                Type:        jsonschema.String,
                Description: "City and country",
            },
        },
        []string{"location"},
        func(params map[string]string) string {
            return `{"temperature":"22°C", "condition":"sunny", "humidity":"60%"}`
        },
    )
    
    // Add currency tool
    agent.AddTool(
        "convert_currency",
        "Convert between currencies",
        map[string]jsonschema.Definition{
            "amount": {Type: jsonschema.String, Description: "Amount to convert"},
            "from":   {Type: jsonschema.String, Description: "Source currency"},
            "to":     {Type: jsonschema.String, Description: "Target currency"},
        },
        []string{"amount", "from", "to"},
        func(params map[string]string) string {
            return fmt.Sprintf(`{"amount":"%s", "from":"%s", "to":"%s", "result":"42.50"}`,
                params["amount"], params["from"], params["to"])
        },
    )
    
    // Set up structured response (optional)
    type Response struct {
        Answer     string `json:"answer"`
        ToolsUsed  []string `json:"tools_used"`
        Confidence float64 `json:"confidence"`
    }
    var response Response
    agent.SetResponseSchema("assistant_response", "Structured response with metadata", true, response)
    
    // Create messages
    message := NewMessages()
    
    // Connect to MCP server for additional capabilities
    err = agent.AddMCP("http://localhost:8080/sse", nil)
    if err != nil {
        log.Printf("MCP server not available: %v", err)
    }
    
    // Ask a question that requires multiple tools (both regular and MCP)
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What's the weather in Paris, how much is 100 USD in EUR, and create a payment link for 50 USD if possible?"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Parse structured response
    if agent.StructuredResponseSchema != nil {
        err = agent.ParseResponse(resp, &response)
        if err != nil {
            log.Printf("Parse error: %v", err)
        } else {
            fmt.Printf("Structured Response: %+v\n", response)
        }
    }
    
    // Or access plain text response
    fmt.Println("Text Response:", resp.Choices[0].Message.Content)
}
```

## Related Documentation

- [LLM Providers](llm.md) - Setting up different language model providers
- [Tools](tool.md) - Detailed tool system and MCP documentation
- [Messages](message.md) - Message creation and management
- [Schemas](schema.md) - JSON schema definitions for structured outputs
- [MCP Specification](https://spec.modelcontextprotocol.io/) - Official Model Context Protocol documentation