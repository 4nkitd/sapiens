# Tools

Tools extend the agent's capabilities by allowing it to call external functions, APIs, or perform specific computations. Sapiens provides a simple yet powerful tool system that integrates seamlessly with all supported LLM providers and includes support for MCP (Model Context Protocol) servers.

## Overview

The tool system in Sapiens allows you to:
- Define custom functions that agents can call
- Connect to MCP servers for external tool integration
- Specify parameter schemas with validation
- Handle tool execution automatically
- Chain multiple tool calls in a single conversation
- Prevent infinite recursion with built-in depth limits
- Seamlessly mix regular tools and MCP tools

## Tool Types

Sapiens supports two types of tools:

### Regular Tools
```go
type AgentTool struct {
    ToolDefinition openai.Tool
    ToolFunction   AgentFunc
}

type AgentFunc func(parameters map[string]string) string
```

Regular tools consist of:
- **ToolDefinition**: OpenAI-compatible tool definition with schema
- **ToolFunction**: Your implementation function that gets called

### MCP Tools
```go
type McpClient struct {
    Ctx       context.Context
    BaseUrl   string
    Client    *mcp_client.Client
    Connected bool
    Tools     []mcp.Tool
}
```

MCP tools are:
- **External tools**: Hosted on MCP servers
- **Protocol-based**: Use Model Context Protocol for communication
- **Auto-discovered**: Tool schemas are fetched from the server
- **Seamlessly integrated**: Work alongside regular tools

## Adding Tools to an Agent

### Regular Tool Addition

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
- `description`: Clear description of what the tool does
- `tool_parameters`: JSON schema definition for parameters
- `required_params`: List of required parameter names
- `funx`: Function that implements the tool logic

### MCP Server Connection

```go
func (a *Agent) AddMCP(url string, customHeaders map[string]string) error
```

**Parameters:**
- `url`: MCP server URL (typically an SSE endpoint like `http://localhost:8080/sse`)
- `customHeaders`: Optional custom headers for authentication or configuration

MCP servers automatically provide their available tools, which are then integrated into the agent.

### Regular Tool Example

```go
import "github.com/sashabaranov/go-openai/jsonschema"

agent.AddTool(
    "get_time",
    "Get the current time",
    map[string]jsonschema.Definition{
        "timezone": {
            Type:        jsonschema.String,
            Description: "Timezone (e.g., UTC, EST, PST)",
        },
    },
    []string{}, // timezone is optional
    func(parameters map[string]string) string {
        timezone := parameters["timezone"]
        if timezone == "" {
            timezone = "UTC"
        }
        
        // Your time logic here
        return fmt.Sprintf(`{"time":"2024-01-15 10:30:00", "timezone":"%s"}`, timezone)
    },
)
```

### MCP Server Example

```go
// Connect to an MCP server that provides payment tools
err := agent.AddMCP("http://localhost:8080/sse", nil)
if err != nil {
    log.Printf("Failed to connect to MCP server: %v", err)
    return
}

fmt.Printf("Connected to MCP server. Agent now has %d MCP tools available\n", len(agent.McpTools))

// The agent can now use MCP tools automatically
// No additional code needed - tools are auto-discovered
```

## Tool Parameter Schemas

Define parameter schemas using JSON Schema definitions:

### Basic Types

```go
// String parameter
"location": {
    Type:        jsonschema.String,
    Description: "City and country name",
}

// Number parameter
"amount": {
    Type:        jsonschema.Number,
    Description: "Amount to convert",
}

// Boolean parameter
"include_forecast": {
    Type:        jsonschema.Boolean,
    Description: "Whether to include forecast data",
}
```

### Enums (Limited Choices)

```go
"unit": {
    Type:        jsonschema.String,
    Enum:        []string{"celsius", "fahrenheit"},
    Description: "Temperature unit",
}

"priority": {
    Type:        jsonschema.String,
    Enum:        []string{"low", "medium", "high", "urgent"},
    Description: "Task priority level",
}
```

### Complex Example

```go
agent.AddTool(
    "book_flight",
    "Book a flight ticket",
    map[string]jsonschema.Definition{
        "from": {
            Type:        jsonschema.String,
            Description: "Departure city/airport code",
        },
        "to": {
            Type:        jsonschema.String,
            Description: "Destination city/airport code",
        },
        "date": {
            Type:        jsonschema.String,
            Description: "Departure date (YYYY-MM-DD)",
        },
        "class": {
            Type:        jsonschema.String,
            Enum:        []string{"economy", "business", "first"},
            Description: "Travel class",
        },
        "passengers": {
            Type:        jsonschema.Number,
            Description: "Number of passengers",
        },
    },
    []string{"from", "to", "date"}, // Required parameters
    func(parameters map[string]string) string {
        from := parameters["from"]
        to := parameters["to"]
        date := parameters["date"]
        class := parameters["class"]
        if class == "" {
            class = "economy"
        }
        
        // Flight booking logic here
        return fmt.Sprintf(`{
            "booking_id": "FL123456",
            "from": "%s",
            "to": "%s",
            "date": "%s",
            "class": "%s",
            "status": "confirmed"
        }`, from, to, date, class)
    },
)
```

## Tool Implementation Functions

Tool functions receive parameters as a map and return a JSON string response.

### Function Signature

```go
type AgentFunc func(parameters map[string]string) string
```

### Best Practices

1. **Always return JSON**: Even for simple responses, return valid JSON
2. **Handle missing parameters**: Check for optional parameters and provide defaults
3. **Error handling**: Return error information in the JSON response
4. **Validation**: Validate input parameters before processing

### Examples

#### Weather Tool

```go
func weatherTool(parameters map[string]string) string {
    location := parameters["location"]
    unit := parameters["unit"]
    if unit == "" {
        unit = "celsius"
    }
    
    // Validate location
    if location == "" {
        return `{"error": "Location is required"}`
    }
    
    // Call weather API (simplified)
    temp := "22"
    if unit == "fahrenheit" {
        temp = "72"
    }
    
    return fmt.Sprintf(`{
        "location": "%s",
        "temperature": "%s",
        "unit": "%s",
        "condition": "sunny",
        "humidity": "65%%",
        "timestamp": "%s"
    }`, location, temp, unit, time.Now().Format(time.RFC3339))
}
```

#### Database Query Tool

```go
func databaseQueryTool(parameters map[string]string) string {
    query := parameters["query"]
    table := parameters["table"]
    
    if query == "" {
        return `{"error": "Query parameter is required"}`
    }
    
    // Sanitize and execute query (implement your DB logic)
    // This is a simplified example
    results := []map[string]interface{}{
        {"id": 1, "name": "John", "email": "john@example.com"},
        {"id": 2, "name": "Jane", "email": "jane@example.com"},
    }
    
    resultJson, _ := json.Marshal(map[string]interface{}{
        "query": query,
        "table": table,
        "results": results,
        "count": len(results),
    })
    
    return string(resultJson)
}
```

#### API Call Tool

```go
func apiCallTool(parameters map[string]string) string {
    endpoint := parameters["endpoint"]
    method := parameters["method"]
    if method == "" {
        method = "GET"
    }
    
    // Make HTTP request (simplified)
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest(method, endpoint, nil)
    if err != nil {
        return fmt.Sprintf(`{"error": "Failed to create request: %s"}`, err.Error())
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Sprintf(`{"error": "Request failed: %s"}`, err.Error())
    }
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    
    return fmt.Sprintf(`{
        "endpoint": "%s",
        "method": "%s",
        "status_code": %d,
        "response": %s
    }`, endpoint, method, resp.StatusCode, string(body))
}
```

## Multiple Tools and MCP Integration

You can add multiple tools to a single agent, including both regular tools and MCP tools. The LLM will automatically choose which tools to use:

```go
// Add regular tools
agent.AddTool("get_weather", "Get weather info", weatherParams, []string{"location"}, weatherFunc)
agent.AddTool("convert_currency", "Convert currency", currencyParams, []string{"amount", "from", "to"}, currencyFunc)
agent.AddTool("get_time", "Get current time", timeParams, []string{}, timeFunc)
agent.AddTool("calculate", "Perform math calculations", calcParams, []string{"expression"}, calcFunc)

// Add MCP server tools
err := agent.AddMCP("http://localhost:8080/sse", nil)
if err != nil {
    log.Printf("MCP server not available: %v", err)
}

// Connect to additional MCP servers
err = agent.AddMCP("http://other-server:9090/sse", map[string]string{
    "Authorization": "Bearer your-token",
})
if err != nil {
    log.Printf("Second MCP server not available: %v", err)
}
```

The agent can then handle complex requests that require multiple tools from different sources:

```go
message := NewMessages()
resp, err := agent.Ask(message.MergeMessages(
    message.UserMessage("What's the weather in Tokyo, what time is it there, convert 100 USD to JPY, and create a payment link for 50 USD"),
))
```

## MCP Tool Discovery and Schema Conversion

When you connect to an MCP server, Sapiens automatically:

1. **Discovers available tools** from the server
2. **Converts MCP schemas** to OpenAI-compatible tool definitions
3. **Integrates tools** seamlessly with existing regular tools
4. **Handles tool calls** transparently

### Schema Conversion Example

```go
// MCP tool schema (automatically fetched from server)
mcpTool := mcp.Tool{
    Name: "createPaymentLink",
    Description: "Create a payment link",
    InputSchema: mcp.ToolInputSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "amount": map[string]interface{}{
                "type": "number",
                "description": "Payment amount",
            },
            "currency": map[string]interface{}{
                "type": "string",
                "enum": []string{"USD", "EUR", "GBP"},
                "description": "Currency code",
            },
        },
        Required: []string{"amount"},
    },
}

// Sapiens automatically converts this to OpenAI format
// No manual conversion needed!
```

## Tool Call Flow

1. **User sends message** containing a request that might need tools
2. **Agent analyzes** the request and determines which tools to call
3. **Tools are executed** with the provided parameters
4. **Tool results** are added to the conversation history
5. **Agent processes** the tool results and generates a final response
6. **Final response** is returned to the user

## Advanced Features

### Recursion Protection

Sapiens automatically prevents infinite tool call loops:
- Maximum depth of 5 tool call levels
- Automatic termination when depth is exceeded
- Error reporting for recursion limits

### Thread Safety

All tool operations are thread-safe and can be used concurrently across multiple goroutines.

### Tool Result Formatting

Tool results are automatically formatted for compatibility with different LLM providers:
- Gemini: Uses user message format for tool responses
- OpenAI: Uses standard tool message format
- Other providers: Automatically adapted

## Error Handling

Handle tool errors gracefully in your implementations:

```go
func robustTool(parameters map[string]string) string {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Tool panic recovered: %v", r)
        }
    }()
    
    // Validate required parameters
    requiredParam := parameters["required_param"]
    if requiredParam == "" {
        return `{"error": "required_param is missing", "code": "MISSING_PARAM"}`
    }
    
    // Perform operation with error handling
    result, err := someOperation(requiredParam)
    if err != nil {
        return fmt.Sprintf(`{"error": "%s", "code": "OPERATION_FAILED"}`, err.Error())
    }
    
    // Return success response
    return fmt.Sprintf(`{"result": "%s", "status": "success"}`, result)
}
```

## Testing Tools

Test your tools independently before adding them to agents:

```go
func TestWeatherTool(t *testing.T) {
    params := map[string]string{
        "location": "London, UK",
        "unit":     "celsius",
    }
    
    result := weatherTool(params)
    
    // Parse JSON response
    var response map[string]interface{}
    err := json.Unmarshal([]byte(result), &response)
    if err != nil {
        t.Fatalf("Invalid JSON response: %v", err)
    }
    
    // Check required fields
    if response["location"] != "London, UK" {
        t.Errorf("Expected location 'London, UK', got %v", response["location"])
    }
    
    if response["unit"] != "celsius" {
        t.Errorf("Expected unit 'celsius', got %v", response["unit"])
    }
}
```

## Complete Example with MCP Integration

Here's a complete example with both regular tools and MCP tools:

```go
package sapiens

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
    // Initialize LLM and agent
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant with access to various tools including MCP tools",
    )
    
    // Add regular weather tool
    agent.AddTool(
        "get_weather",
        "Get current weather for a location",
        map[string]jsonschema.Definition{
            "location": {Type: jsonschema.String, Description: "City and country"},
            "unit":     {Type: jsonschema.String, Enum: []string{"celsius", "fahrenheit"}},
        },
        []string{"location"},
        func(params map[string]string) string {
            return fmt.Sprintf(`{"temperature":"25", "condition":"sunny", "location":"%s"}`, params["location"])
        },
    )
    
    // Add regular time tool
    agent.AddTool(
        "get_current_time",
        "Get current time in a timezone",
        map[string]jsonschema.Definition{
            "timezone": {Type: jsonschema.String, Description: "Timezone (UTC, EST, PST, etc.)"},
        },
        []string{},
        func(params map[string]string) string {
            timezone := params["timezone"]
            if timezone == "" {
                timezone = "UTC"
            }
            return fmt.Sprintf(`{"time":"%s", "timezone":"%s"}`, time.Now().Format("15:04:05"), timezone)
        },
    )
    
    // Connect to MCP server for additional tools (e.g., payment tools)
    err := agent.AddMCP("http://localhost:8080/sse", nil)
    if err != nil {
        fmt.Printf("Warning: Could not connect to MCP server: %v\n", err)
        fmt.Println("Continuing with regular tools only...")
    } else {
        fmt.Printf("Successfully connected to MCP server with %d additional tools\n", len(agent.McpTools))
    }
    
    // Display available tools
    fmt.Printf("Agent has %d regular tools and %d MCP tools\n", 
        len(agent.Tools), len(agent.McpTools))
    
    // Use the agent with mixed tool capabilities
    message := NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("What's the weather in Paris, what time is it there, and create a payment link for 100 USD if possible?"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

## MCP Server Setup

To use MCP tools, you need an MCP server running. Here's a basic example of setting up a simple MCP server:

### Prerequisites
1. An MCP server implementation (various implementations available)
2. Server running on accessible URL (e.g., `http://localhost:8080/sse`)
3. Server exposing tools via MCP protocol

### Testing MCP Connection
```go
// Test MCP connection and list available tools
mcpClient, err := NewMcpClient(context.Background(), "http://localhost:8080/sse")
if err != nil {
    log.Printf("MCP connection failed: %v", err)
    return
}
defer mcpClient.Disconnect()

// List available tools
toolsResult, err := mcpClient.ListTools()
if err != nil {
    log.Printf("Failed to list MCP tools: %v", err)
    return
}

fmt.Printf("Found %d MCP tools:\n", len(toolsResult.Tools))
for _, tool := range toolsResult.Tools {
    fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
}
```

## MCP Error Handling

MCP connections can fail for various reasons. Handle them gracefully:

```go
// Attempt MCP connection with fallback
err := agent.AddMCP("http://localhost:8080/sse", nil)
if err != nil {
    log.Printf("Primary MCP server unavailable: %v", err)
    
    // Try backup server
    err = agent.AddMCP("http://backup-server:8080/sse", nil)
    if err != nil {
        log.Printf("Backup MCP server also unavailable: %v", err)
        log.Println("Continuing with regular tools only")
    } else {
        fmt.Println("Connected to backup MCP server")
    }
} else {
    fmt.Println("Connected to primary MCP server")
}
```

## MCP Tool Call Flow

1. **Agent analyzes** user request and identifies needed tools
2. **Tool selection** happens across both regular and MCP tools
3. **Regular tools** are executed locally via function calls
4. **MCP tools** are executed via protocol calls to MCP server
5. **Results are merged** and processed together
6. **Final response** incorporates results from all tool types

## Advanced MCP Features

### Custom Headers for Authentication
```go
// Add custom headers for authenticated MCP servers
headers := map[string]string{
    "Authorization": "Bearer your-jwt-token",
    "X-API-Key":     "your-api-key",
    "Content-Type":  "application/json",
}

err := agent.AddMCP("https://secure-mcp-server.com/sse", headers)
```

### Multiple MCP Servers
```go
// Connect to multiple MCP servers for different capabilities
agent.AddMCP("http://payments-server:8080/sse", nil)      // Payment tools
agent.AddMCP("http://analytics-server:9090/sse", nil)     // Analytics tools
agent.AddMCP("http://notifications-server:7070/sse", nil) // Notification tools

// Agent automatically aggregates tools from all servers
```

### MCP Tool Inspection
```go
// Inspect MCP tools after connection
for _, mcpTool := range agent.McpTools {
    fmt.Printf("MCP Tool: %s\n", mcpTool.Name)
    fmt.Printf("Description: %s\n", mcpTool.Description)
    fmt.Printf("Schema: %+v\n", mcpTool.InputSchema)
}
```

## Related Documentation

- [Agent](agent.md) - Using tools with agents and MCP integration
- [Schema](schema.md) - JSON schema definitions for tool parameters
- [Examples](examples.md) - More tool examples and MCP patterns
- [MCP Specification](https://spec.modelcontextprotocol.io/) - Official MCP documentation