# MCP (Model Context Protocol) Setup Guide

This guide explains how to set up and use MCP (Model Context Protocol) servers with Sapiens for extended agent capabilities.

## What is MCP?

Model Context Protocol (MCP) is a standardized protocol that allows AI agents to securely connect to external data sources and tools. It enables agents to access real-time information, perform actions, and integrate with various services without requiring custom integrations for each tool.

## Benefits of MCP Integration

- **External Tool Access**: Connect to services like payment processors, databases, APIs
- **Real-time Data**: Access live data from external sources
- **Standardized Protocol**: Use any MCP-compatible server
- **Security**: Built-in authentication and authorization
- **Scalability**: Add multiple MCP servers for different capabilities

## Prerequisites

1. **Go Environment**: Ensure you have Go 1.19+ installed
2. **Sapiens Library**: Install with `go get github.com/4nkitd/sapiens`
3. **MCP Server**: A running MCP server (see setup options below)

## MCP Server Options

### Option 1: Test MCP Server (Recommended for Development)

Create a simple test MCP server for development:

```bash
# Create a new directory for your MCP server
mkdir mcp-test-server
cd mcp-test-server

# Initialize Go module
go mod init mcp-test-server

# Create main.go with basic MCP server
```

Example test server (`main.go`):

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/mark3labs/mcp-go/server/transport"
)

func main() {
    // Create MCP server
    mcpServer := server.NewServer("test-server", "1.0.0")
    
    // Add a simple payment tool
    mcpServer.AddTool(mcp.Tool{
        Name:        "createPaymentLink",
        Description: "Create a payment link for a specified amount",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "amount": map[string]interface{}{
                    "type":        "number",
                    "description": "Payment amount in USD",
                },
                "description": map[string]interface{}{
                    "type":        "string",
                    "description": "Payment description",
                },
                "email": map[string]interface{}{
                    "type":        "string",
                    "description": "Customer email address",
                },
            },
            Required: []string{"amount", "email"},
        },
    }, func(ctx context.Context, params mcp.CallToolParams) (*mcp.CallToolResult, error) {
        amount := params.Arguments["amount"]
        description := params.Arguments["description"]
        email := params.Arguments["email"]
        
        // Simulate payment link creation
        paymentID := fmt.Sprintf("pay_%d", time.Now().Unix())
        
        return &mcp.CallToolResult{
            Content: []interface{}{
                map[string]interface{}{
                    "type": "text",
                    "text": fmt.Sprintf(`{
                        "payment_id": "%s",
                        "amount": %v,
                        "description": "%v",
                        "email": "%v",
                        "payment_url": "https://pay.example.com/%s",
                        "status": "created",
                        "created_at": "%s"
                    }`, paymentID, amount, description, email, paymentID, time.Now().Format(time.RFC3339)),
                },
            },
        }, nil
    })
    
    // Add another tool for fetching payment status
    mcpServer.AddTool(mcp.Tool{
        Name:        "getPaymentStatus",
        Description: "Get the status of a payment by ID",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "payment_id": map[string]interface{}{
                    "type":        "string",
                    "description": "Payment ID to check",
                },
            },
            Required: []string{"payment_id"},
        },
    }, func(ctx context.Context, params mcp.CallToolParams) (*mcp.CallToolResult, error) {
        paymentID := params.Arguments["payment_id"]
        
        // Simulate payment status check
        status := "completed"
        if paymentID == "" {
            status = "not_found"
        }
        
        return &mcp.CallToolResult{
            Content: []interface{}{
                map[string]interface{}{
                    "type": "text",
                    "text": fmt.Sprintf(`{
                        "payment_id": "%v",
                        "status": "%s",
                        "updated_at": "%s"
                    }`, paymentID, status, time.Now().Format(time.RFC3339)),
                },
            },
        }, nil
    })
    
    // Create SSE transport
    transport := transport.NewSSE()
    
    // Create HTTP handler
    handler := func(w http.ResponseWriter, r *http.Request) {
        transport.HandleSSE(w, r, mcpServer)
    }
    
    // Start HTTP server
    http.HandleFunc("/sse", handler)
    
    fmt.Println("MCP test server starting on http://localhost:8080")
    fmt.Println("SSE endpoint: http://localhost:8080/sse")
    fmt.Println("Available tools: createPaymentLink, getPaymentStatus")
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Install dependencies and run:

```bash
go mod tidy
go run main.go
```

### Option 2: Docker MCP Server

Use a pre-built Docker container:

```bash
# Pull and run a sample MCP server
docker run -p 8080:8080 mcpserver/sample-server
```

### Option 3: Production MCP Servers

For production use, consider these MCP server implementations:

1. **Python MCP Server**: https://github.com/modelcontextprotocol/python-sdk
2. **Node.js MCP Server**: https://github.com/modelcontextprotocol/node-sdk
3. **Rust MCP Server**: https://github.com/modelcontextprotocol/rust-sdk

## Basic Sapiens + MCP Integration

### 1. Simple Connection

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/4nkitd/sapiens"
)

func main() {
    // Initialize Sapiens agent
    llm := sapiens.NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := sapiens.NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant with payment capabilities",
    )
    
    // Connect to MCP server
    err := agent.AddMCP("http://localhost:8080/sse", nil)
    if err != nil {
        log.Fatalf("Failed to connect to MCP server: %v", err)
    }
    
    fmt.Printf("Connected! Agent has %d MCP tools available\n", len(agent.McpTools))
    
    // Use MCP tools
    message := sapiens.NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Create a payment link for $50 from customer@example.com"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### 2. With Authentication

```go
// For authenticated MCP servers
headers := map[string]string{
    "Authorization": "Bearer your-jwt-token",
    "X-API-Key":     "your-api-key",
}

err := agent.AddMCP("https://secure-mcp-server.com/sse", headers)
if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}
```

### 3. Multiple MCP Servers

```go
// Connect to multiple specialized servers
agent.AddMCP("http://payments-mcp:8080/sse", nil)        // Payment tools
agent.AddMCP("http://analytics-mcp:9090/sse", nil)       // Analytics tools
agent.AddMCP("http://notifications-mcp:7070/sse", nil)   // Notification tools

fmt.Printf("Total MCP tools: %d\n", len(agent.McpTools))
```

## Environment Setup

### Development Environment

Create a `.env` file:

```bash
# LLM Provider API Keys
GEMINI_API_KEY=your-gemini-key
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key

# MCP Server URLs
MCP_PAYMENTS_URL=http://localhost:8080/sse
MCP_ANALYTICS_URL=http://localhost:9090/sse

# MCP Authentication (if needed)
MCP_TOKEN=your-mcp-token
MCP_API_KEY=your-mcp-api-key
```

### Production Environment

For production deployments:

1. **Use HTTPS**: Ensure MCP servers use HTTPS endpoints
2. **Authentication**: Implement proper API key or JWT authentication
3. **Rate Limiting**: Configure appropriate rate limits
4. **Monitoring**: Set up health checks and monitoring
5. **Error Handling**: Implement robust error handling and fallbacks

## Common MCP Server Setups

### Payment Processing Server

Example tools a payment MCP server might provide:

- `createPaymentLink` - Generate payment links
- `processPayment` - Process direct payments
- `getPaymentStatus` - Check payment status
- `refundPayment` - Process refunds
- `listTransactions` - List transaction history

### Database Server

Example tools for database operations:

- `executeQuery` - Run SQL queries
- `insertRecord` - Insert new records
- `updateRecord` - Update existing records
- `deleteRecord` - Delete records
- `getSchema` - Retrieve table schemas

### Analytics Server

Example tools for analytics:

- `getMetrics` - Retrieve key metrics
- `runReport` - Generate reports
- `createDashboard` - Create dashboards
- `exportData` - Export data sets
- `getInsights` - AI-powered insights

## Testing Your MCP Setup

### 1. Test MCP Connection

```go
func testMCPConnection() {
    mcpClient, err := sapiens.NewMcpClient(context.Background(), "http://localhost:8080/sse")
    if err != nil {
        log.Printf("Connection failed: %v", err)
        return
    }
    defer mcpClient.Disconnect()
    
    // List available tools
    tools, err := mcpClient.ListTools()
    if err != nil {
        log.Printf("Failed to list tools: %v", err)
        return
    }
    
    fmt.Printf("Found %d tools:\n", len(tools.Tools))
    for _, tool := range tools.Tools {
        fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
    }
}
```

### 2. Test Tool Calls

```go
func testToolCall() {
    agent := /* ... initialize agent ... */
    
    // Test each MCP tool
    for _, tool := range agent.McpTools {
        fmt.Printf("Testing tool: %s\n", tool.Name)
        
        message := sapiens.NewMessages()
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(fmt.Sprintf("Please test the %s tool with sample data", tool.Name)),
        ))
        
        if err != nil {
            log.Printf("Tool test failed: %v", err)
        } else {
            fmt.Printf("Tool response: %s\n", resp.Choices[0].Message.Content)
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   ```
   Error: dial tcp [::1]:8080: connect: connection refused
   ```
   **Solution**: Ensure MCP server is running on the specified port

2. **Invalid SSE Endpoint**
   ```
   Error: error creating MCP server transport
   ```
   **Solution**: Verify the SSE endpoint URL (usually ends with `/sse`)

3. **Authentication Failed**
   ```
   Error: 401 Unauthorized
   ```
   **Solution**: Check authentication headers and credentials

4. **Tool Schema Errors**
   ```
   Error: failed to parse MCP tool arguments
   ```
   **Solution**: Verify tool parameter schemas match expected format

### Debug Mode

Enable debug logging in your MCP client:

```go
// The MCP client includes debug logging by default
// Check console output for detailed connection and tool call information
```

### Health Checks

Implement health checks for MCP servers:

```go
func checkMCPHealth(url string) bool {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    mcpClient, err := sapiens.NewMcpClient(ctx, url)
    if err != nil {
        return false
    }
    defer mcpClient.Disconnect()
    
    _, err = mcpClient.ListTools()
    return err == nil
}
```

## Best Practices

### 1. Error Handling

Always implement graceful error handling:

```go
err := agent.AddMCP(mcpURL, nil)
if err != nil {
    log.Printf("MCP server unavailable: %v", err)
    log.Println("Continuing with regular tools only...")
    // Continue execution with reduced functionality
}
```

### 2. Fallback Tools

Provide fallback tools for when MCP servers are unavailable:

```go
// Add regular tools as fallbacks
agent.AddTool("basic_payment_info", "Basic payment info when service unavailable", ...)

// Then try to add MCP tools
err := agent.AddMCP(mcpURL, nil)
if err != nil {
    log.Println("Using fallback tools")
}
```

### 3. Connection Pooling

For high-traffic applications, consider connection pooling:

```go
// Reuse agent instances instead of creating new ones for each request
var globalAgent *sapiens.Agent

func init() {
    globalAgent = sapiens.NewAgent(...)
    globalAgent.AddMCP(mcpURL, nil)
}
```

### 4. Security

- Use HTTPS for production MCP servers
- Implement proper authentication
- Validate all tool inputs
- Use environment variables for sensitive data

### 5. Monitoring

Monitor MCP server health and performance:

```go
func monitorMCPServers() {
    servers := []string{
        "http://payments-mcp:8080/sse",
        "http://analytics-mcp:9090/sse",
    }
    
    for _, server := range servers {
        if !checkMCPHealth(server) {
            log.Printf("MCP server unhealthy: %s", server)
            // Implement alerting logic
        }
    }
}
```

## Next Steps

1. **Start with the test server** provided in this guide
2. **Experiment with different tools** and see how they integrate
3. **Build your own MCP server** for custom functionality
4. **Implement error handling and fallbacks** for production use
5. **Scale up** with multiple MCP servers for different capabilities

## Resources

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [MCP Go Library](https://github.com/mark3labs/mcp-go)
- [Sapiens Documentation](https://github.com/4nkitd/sapiens)
- [Example MCP Servers](https://github.com/modelcontextprotocol)

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Verify your MCP server is running and accessible
3. Review the debug logs for connection details
4. Open an issue on the Sapiens GitHub repository

Happy coding with MCP and Sapiens! 🚀