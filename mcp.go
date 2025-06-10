package sapiens

import (
	"context"
	"fmt"
	"time"

	mcp_client "github.com/mark3labs/mcp-go/client"
	mcp_transport "github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type McpClient struct {
	Ctx       context.Context
	BaseUrl   string
	Client    *mcp_client.Client
	Connected bool
	Tools     []mcp.Tool
}

func NewMcpClient(ctx context.Context, mcp_sse_url string) (*McpClient, error) {
	mcp_server_transport, mcp_server_transport_err := mcp_transport.NewSSE(mcp_sse_url)
	if mcp_server_transport_err != nil {
		return nil, fmt.Errorf("error creating MCP server transport: %w", mcp_server_transport_err)
	}
	
	mcp_client_instance := mcp_client.NewClient(mcp_server_transport)

	// Start the client with timeout
	startCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	if err := mcp_client_instance.Start(startCtx); err != nil {
		return nil, fmt.Errorf("error starting MCP client: %w", err)
	}

	// Initialize with timeout
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	_, err := mcp_client_instance.Initialize(initCtx, mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("error initializing MCP client: %w", err)
	}

	mcpClient := &McpClient{
		BaseUrl:   mcp_sse_url,
		Client:    mcp_client_instance,
		Ctx:       ctx,
		Connected: true,
	}

	// Cache available tools
	if err := mcpClient.refreshTools(); err != nil {
		fmt.Printf("Warning: could not load MCP tools: %v\n", err)
	}

	return mcpClient, nil
}

func (m *McpClient) ListTools() (*mcp.ListToolsResult, error) {
	if !m.Connected {
		return nil, fmt.Errorf("MCP client is not connected")
	}

	listCtx, cancel := context.WithTimeout(m.Ctx, 5*time.Second)
	defer cancel()

	listToolsResult, listToolsResultErr := m.Client.ListTools(listCtx, mcp.ListToolsRequest{})
	if listToolsResultErr != nil {
		m.Connected = false
		return nil, fmt.Errorf("error listing MCP tools: %w", listToolsResultErr)
	}

	return listToolsResult, listToolsResultErr
}

func (m *McpClient) CallTool(request mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if !m.Connected {
		return nil, fmt.Errorf("MCP client is not connected")
	}

	callCtx, cancel := context.WithTimeout(m.Ctx, 30*time.Second)
	defer cancel()

	callToolResult, callToolResultErr := m.Client.CallTool(callCtx, mcp.CallToolRequest{
		Params: request,
	})

	if callToolResultErr != nil {
		return nil, fmt.Errorf("error calling MCP tool '%s': %w", request.Name, callToolResultErr)
	}

	return callToolResult, callToolResultErr
}

func (m *McpClient) refreshTools() error {
	toolsResult, err := m.ListTools()
	if err != nil {
		return err
	}
	
	m.Tools = toolsResult.Tools
	return nil
}

func (m *McpClient) GetCachedTools() []mcp.Tool {
	return m.Tools
}

func (m *McpClient) IsConnected() bool {
	return m.Connected
}

func (m *McpClient) Disconnect() error {
	if m.Client != nil {
		m.Connected = false
		// Note: The mcp-go library doesn't seem to have a Close() method
		// This would be where we'd close the connection if available
	}
	return nil
}

func (m *McpClient) HasTool(toolName string) bool {
	for _, tool := range m.Tools {
		if tool.Name == toolName {
			return true
		}
	}
	return false
}
