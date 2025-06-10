package sapiens

import (
	"context"
	"fmt"
	"testing"

	mcp_client "github.com/mark3labs/mcp-go/client"
	mcp_transport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
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
}
