# Sapiens API Documentation

This document provides detailed information about the Sapiens API.

## Agent

The `Agent` is the core interface for interacting with AI models and managing tools including MCP (Model Context Protocol) integration.

### Creating a New Agent

```go
agent := NewAgent(ctx context.Context, llm *openai.Client, model string, systemPrompt string) *Agent
```

**Parameters:**
- `ctx`: Context for operations and cancellation
- `llm`: OpenAI-compatible client from any provider
- `model`: Model name (e.g., "gpt-4", "claude-sonet-3.5", "gemini-2.0-flash")
- `systemPrompt`: System prompt defining agent behavior

**Example:**
```go
llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
agent := NewAgent(
    context.Background(),
    llm.Client(),
    llm.GetDefaultModel(),
    "You are a helpful assistant",
)
```

### Methods

#### Ask

```go
Ask(messages []ChatCompletionMessage) (ChatCompletionResponse, error)
```

Sends messages to the agent and returns the complete response.

**Parameters:**
- `messages`: Array of chat messages created using the Messages helper

**Example:**
```go
message := NewMessages()
resp, err := agent.Ask(message.MergeMessages(
    message.UserMessage("Hello, how are you?"),
))
```

#### AddTool

```go
AddTool(name, description string, parameters map[string]jsonschema.Definition, required []string, callback AgentFunc) error
```

Adds a regular tool to the agent.

**Parameters:**
- `name`: Unique tool identifier
- `description`: Tool description
- `parameters`: JSON schema for tool parameters
- `required`: List of required parameter names
- `callback`: Function implementing tool logic

#### AddMCP

```go
AddMCP(url string, customHeaders map[string]string) error
```

Connects to an MCP server and adds its tools to the agent.

**Parameters:**
- `url`: MCP server URL (typically SSE endpoint)
- `customHeaders`: Optional authentication headers

**Example:**
```go
// Basic connection
err := agent.AddMCP("http://localhost:8080/sse", nil)

// With authentication
headers := map[string]string{
    "Authorization": "Bearer token",
}
err := agent.AddMCP("https://secure-mcp.com/sse", headers)
```

#### SetResponseSchema

```go
SetResponseSchema(name, description string, strict bool, schema interface{}) *ChatCompletionResponseFormat
```

Configures structured output schema for responses.

#### ParseResponse

```go
ParseResponse(response ChatCompletionResponse, target interface{}) error
```

Parses a structured response into a Go struct.

#### GetMcpToolByName

```go
GetMcpToolByName(name string) (mcp.Tool, error)
```

Retrieves an MCP tool by name for inspection.

## Tools

Sapiens supports two types of tools: regular tools (local functions) and MCP tools (external services).

### Regular Tools

Created using `AddTool()` method with local function implementations.

**Tool Function Signature:**
```go
type AgentFunc func(parameters map[string]string) string
```

**Example:**
```go
agent.AddTool(
    "get_weather",
    "Get weather information",
    map[string]jsonschema.Definition{
        "location": {
            Type:        jsonschema.String,
            Description: "City name",
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
        return fmt.Sprintf(`{"temperature":"25", "unit":"%s", "location":"%s"}`, unit, location)
    },
)
```

### MCP Tools

External tools from MCP servers, automatically discovered and integrated.

**Features:**
- Auto-discovery from MCP servers
- Automatic schema conversion
- Seamless integration with regular tools
- Support for multiple MCP servers

**Connection:**
```go
err := agent.AddMCP("http://mcp-server:8080/sse", nil)
```

### Tool Integration

Both tool types work together seamlessly:
- Agent automatically chooses appropriate tools
- Mixed tool calls in single conversations
- Consistent error handling and recursion protection

### Tool Structure

**Regular Tool:**
```go
type AgentTool struct {
    ToolDefinition openai.Tool
    ToolFunction   AgentFunc
}
```

**MCP Tool:**
```go
type McpClient struct {
    Ctx       context.Context
    BaseUrl   string
    Client    *mcp_client.Client
    Connected bool
    Tools     []mcp.Tool
}
```

## Message Management

Sapiens provides utilities for creating and managing conversation messages.

### Messages Helper

```go
type Messages struct{}

func NewMessages() *Messages
```

**Methods:**
- `UserMessage(content string) ChatCompletionMessage`
- `AgentMessage(content string) ChatCompletionMessage`
- `ToolMessage(id, name, content string) ChatCompletionMessage`
- `MergeMessages(...messages) []ChatCompletionMessage`

**Example:**
```go
message := NewMessages()
messages := message.MergeMessages(
    message.UserMessage("Hello"),
    message.AgentMessage("Hi there!"),
)
```

### Conversation History

The agent automatically maintains conversation history:
- All messages stored in `MessagesHistory`
- Context preserved across interactions
- Tool calls and responses included
- Thread-safe access with mutex protection

## MCP Client API

Direct MCP client interface for advanced usage.

### Creating MCP Client

```go
NewMcpClient(ctx context.Context, mcpURL string) (*McpClient, error)
```

### MCP Client Methods

#### ListTools

```go
ListTools() (*mcp.ListToolsResult, error)
```

Lists available tools from the MCP server.

#### CallTool

```go
CallTool(request mcp.CallToolParams) (*mcp.CallToolResult, error)
```

Calls an MCP tool directly.

#### Schema Conversion

```go
ParseToolDefinition(tool mcp.ToolInputSchema) map[string]jsonschema.Definition
EncodeToolDefinition(tool map[string]jsonschema.Definition) mcp.ToolInputSchema
```

Converts between MCP and OpenAI tool schemas.

#### Connection Management

```go
IsConnected() bool
Disconnect() error
HasTool(toolName string) bool
GetCachedTools() []mcp.Tool
```

## LLM Providers

Sapiens supports multiple LLM providers with unified interface.

### OpenAI

```go
llm := NewOpenai(apiKey string)
// Default model: gpt-4.1-2025-04-14
```

### Google Gemini

```go
llm := NewGemini(apiKey string)
// Default model: gemini-2.0-flash
```

### Anthropic Claude

```go
llm := NewAnthropic(apiKey string)
// Default model: claude-sonet-3.5
```

### Ollama

```go
llm := NewOllama(baseURL, authToken, modelName string)
```

## Error Handling

Common error patterns:

- **Tool Errors**: Tool execution failures return descriptive errors
- **MCP Connection Errors**: Network or authentication failures
- **Schema Errors**: Invalid tool parameter schemas
- **Recursion Errors**: Maximum tool call depth exceeded (limit: 5)
- **Provider Errors**: LLM provider API errors

**Error Handling Example:**
```go
resp, err := agent.Ask(messages)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "MCP"):
        log.Printf("MCP error: %v", err)
    case strings.Contains(err.Error(), "tool call"):
        log.Printf("Tool error: %v", err)
    default:
        log.Printf("General error: %v", err)
    }
}
```

## Thread Safety

All Sapiens operations are thread-safe:
- Agent methods protected by mutexes
- Safe concurrent tool calls
- Protected message history access
- Safe MCP client operations
