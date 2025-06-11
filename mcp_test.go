package sapiens

import (
	"context"
	"fmt"
	"testing"
	"time"

	mcp_client "github.com/mark3labs/mcp-go/client"
	mcp_transport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestMCP(t *testing.T) {

	mcp_sse_url := "http://localhost:8080/sse"

	mcp_server_transport, mcp_server_transport_err := mcp_transport.NewSSE(mcp_sse_url)
	if mcp_server_transport_err != nil {
		fmt.Println("Error creating MCP server transport:", mcp_server_transport_err)
	}
	mcp_client_instamce := mcp_client.NewClient(mcp_server_transport)

	mcp_client_instamce.Start(context.Background())

	// resp, err :=
	mcp_client_instamce.Initialize(context.Background(), mcp.InitializeRequest{})
	// fmt.Printf("DEBUG:: %+v  , %+v", resp, err)

	respTool, errTool := mcp_client_instamce.ListTools(context.Background(), mcp.ListToolsRequest{})
	fmt.Printf("DEBUG:: %+v  , %+v", respTool, errTool)

	// Create sample arguments based on the tool schema
	sampleArgs := map[string]interface{}{
		"amount":       100,
		"receipt_id":   fmt.Sprintf("test-receipt-%d", time.Now().Unix()),
		"callback_url": "https://example.com/callback",
		"cname":        "Test Customer",
		"phone":        "+1234567890",
	}

	respt, resptErr := mcp_client_instamce.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "createOrder",
			Arguments: sampleArgs,
		},
	})

	fmt.Printf("t Call Resp:: %+v %+v", respt, resptErr)

}

func TestMcpFull(t *testing.T) {
	// Create a context with a longer timeout for the entire test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cl, clErr := NewMcpClient(ctx, "http://localhost:8080/sse")
	if clErr != nil {
		t.Fatalf("Error creating MCP client: %v", clErr)
	}

	// Ensure client is properly disconnected at the end
	defer func() {
		if err := cl.Disconnect(); err != nil {
			fmt.Printf("Warning: Error disconnecting MCP client: %v\n", err)
		}
	}()

	// Verify client is connected
	if !cl.IsConnected() {
		t.Fatal("MCP client should be connected after creation")
	}

	tools, toolErr := cl.ListTools()
	if toolErr != nil {
		t.Fatalf("Error listing tools: %v", toolErr)
	}

	if len(tools.Tools) == 0 {
		t.Skip("No tools available from MCP server, skipping tool call tests")
	}

	toolsCalls := []map[string]jsonschema.Definition{}

	for i, tool := range tools.Tools {
		fmt.Printf("Testing tool %d/%d: %s\n", i+1, len(tools.Tools), tool.Name)
		fmt.Println("Tool InputSchema:", tool.InputSchema)

		parsedDefinitions := cl.ParseToolDefinition(tool.InputSchema)
		toolsCalls = append(toolsCalls, parsedDefinitions)

		// Test the round-trip conversion: ParseToolDefinition -> EncodeToolDefinition
		if len(parsedDefinitions) > 0 {
			fmt.Printf("Parsed %d properties for tool %s\n", len(parsedDefinitions), tool.Name)

			// Test encoding the parsed definitions back to ToolInputSchema
			encodedSchema := cl.EncodeToolDefinition(parsedDefinitions)
			fmt.Printf("Encoded schema type: %s, properties count: %d\n",
				encodedSchema.Type, len(encodedSchema.Properties))

			// Create sample arguments based on the tool schema
			var sampleArgs map[string]interface{}
			if tool.Name == "createOrder" {
				sampleArgs = map[string]interface{}{
					"amount":       100,
					"receipt_id":   fmt.Sprintf("test-receipt-%d", time.Now().Unix()),
					"callback_url": "https://example.com/callback",
					"cname":        "Test Customer",
					"phone":        "+1234567890",
				}
			} else if tool.Name == "fetchOrder" {
				sampleArgs = map[string]interface{}{
					"order_id": "test-order-123",
				}
			} else {
				// For other tools, create empty args or skip
				fmt.Printf("Skipping tool call for unknown tool: %s\n", tool.Name)
				continue
			}

			fmt.Printf("Calling tool '%s' with args: %+v\n", tool.Name, sampleArgs)

			// Add retry logic for individual tool calls
			var toolResult *mcp.CallToolResult
			var toolcallErr error
			maxRetries := 3

			for retry := 0; retry < maxRetries; retry++ {
				if retry > 0 {
					fmt.Printf("Retrying tool call %d/%d...\n", retry+1, maxRetries)
					time.Sleep(time.Duration(retry) * 2 * time.Second)
				}

				toolResult, toolcallErr = cl.CallTool(mcp.CallToolParams{
					Name:      tool.Name,
					Arguments: sampleArgs,
				})

				if toolcallErr == nil {
					fmt.Printf("Tool call successful for '%s'\n", tool.Name)
					break
				}

				fmt.Printf("Tool call failed (attempt %d/%d): %v\n", retry+1, maxRetries, toolcallErr)
			}

			if toolcallErr != nil {
				fmt.Printf("All retries failed for tool '%s': %v\n", tool.Name, toolcallErr)
			} else {
				fmt.Printf("Tool result for '%s': %+v\n", tool.Name, toolResult)
			}
		}
	}

	fmt.Printf("Completed testing %d tools\n", len(tools.Tools))
}
