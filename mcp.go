package sapiens

import (
	"context"
	"fmt"

	mcp_client "github.com/mark3labs/mcp-go/client"
	mcp_transport "github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type McpClient struct {
	Ctx     context.Context
	BaseUrl string
	Client  *mcp_client.Client
}

func NewMcpClient(ctx context.Context, mcp_sse_url string) *McpClient {

	mcp_server_transport, mcp_server_transport_err := mcp_transport.NewSSE(mcp_sse_url)
	if mcp_server_transport_err != nil {
		fmt.Println("Error creating MCP server transport:", mcp_server_transport_err)
	}
	mcp_client_instamce := mcp_client.NewClient(mcp_server_transport)

	mcp_client_instamce.Start(context.Background())

	// resp, err :=
	mcp_client_instamce.Initialize(context.Background(), mcp.InitializeRequest{})

	return &McpClient{
		BaseUrl: mcp_sse_url,
		Client:  mcp_client_instamce,
		Ctx:     ctx,
	}

}

func (m *McpClient) ListTools() (*mcp.ListToolsResult, error) {
	listToolsResult, listToolsResultErr := m.Client.ListTools(m.Ctx, mcp.ListToolsRequest{})

	return listToolsResult, listToolsResultErr
}

func (m *McpClient) CallTool(request mcp.CallToolParams) (*mcp.CallToolResult, error) {

	listToolsResult, listToolsResultErr := m.Client.CallTool(m.Ctx, mcp.CallToolRequest{
		Params: request,
	})

	return listToolsResult, listToolsResultErr

}
