package sapiens

import (
	"context"
	"fmt"

	mcp_client "github.com/mark3labs/mcp-go/client"
	mcp_transport "github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type McpClient struct {
	Ctx       context.Context
	BaseUrl   string
	Client    *mcp_client.Client
	Connected bool
	Tools     []mcp.Tool
}

func NewMcpClient(ctx context.Context, mcp_sse_url string) (*McpClient, error) {
	fmt.Printf("DEBUG: Creating MCP client for URL: %s\n", mcp_sse_url)
	
	mcp_server_transport, mcp_server_transport_err := mcp_transport.NewSSE(mcp_sse_url)
	if mcp_server_transport_err != nil {
		return nil, fmt.Errorf("error creating MCP server transport: %w", mcp_server_transport_err)
	}
	fmt.Printf("DEBUG: MCP transport created successfully\n")

	mcp_client_instance := mcp_client.NewClient(mcp_server_transport)
	fmt.Printf("DEBUG: MCP client instance created\n")

	fmt.Printf("DEBUG: Starting MCP client...\n")
	if err := mcp_client_instance.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("error starting MCP client: %w", err)
	}
	fmt.Printf("DEBUG: MCP client started successfully\n")

	fmt.Printf("DEBUG: Initializing MCP client...\n")
	initResp, err := mcp_client_instance.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("error initializing MCP client: %w", err)
	}
	fmt.Printf("DEBUG: MCP client initialized successfully. Response: %+v\n", initResp)

	mcpClient := &McpClient{
		BaseUrl:   mcp_sse_url,
		Client:    mcp_client_instance,
		Ctx:       context.Background(),
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

	if len(m.Tools) > 1 {
		return &mcp.ListToolsResult{
			Tools: m.Tools,
		}, nil
	}

	listToolsResult, listToolsResultErr := m.Client.ListTools(context.Background(), mcp.ListToolsRequest{})
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

	fmt.Printf("DEBUG: Calling MCP tool '%s' with args: %+v\n", request.Name, request.Arguments)

	callToolResult, callToolResultErr := m.Client.CallTool(context.Background(), mcp.CallToolRequest{
		Params: request,
	})

	if callToolResultErr != nil {
		fmt.Printf("DEBUG: MCP tool call error: %v\n", callToolResultErr)
		return nil, fmt.Errorf("error calling MCP tool '%s': %w", request.Name, callToolResultErr)
	}

	fmt.Printf("DEBUG: MCP tool call successful. Result: %+v\n", callToolResult)
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

func (m *McpClient) EncodeToolDefinition(tool map[string]jsonschema.Definition) mcp.ToolInputSchema {
	properties := make(map[string]any)
	var required []string

	// Convert each jsonschema.Definition to MCP tool schema format
	for propName, definition := range tool {
		propMap := make(map[string]interface{})

		// Set type
		switch definition.Type {
		case jsonschema.String:
			propMap["type"] = "string"
		case jsonschema.Object:
			propMap["type"] = "object"
		case jsonschema.Number:
			propMap["type"] = "number"
		case jsonschema.Integer:
			propMap["type"] = "integer"
		case jsonschema.Boolean:
			propMap["type"] = "boolean"
		case jsonschema.Array:
			propMap["type"] = "array"
		default:
			propMap["type"] = "string"
		}

		// Set description if present
		if definition.Description != "" {
			propMap["description"] = definition.Description
		}

		// Set enum values if present
		if len(definition.Enum) > 0 {
			enumInterface := make([]interface{}, len(definition.Enum))
			for i, v := range definition.Enum {
				enumInterface[i] = v
			}
			propMap["enum"] = enumInterface
		}

		properties[propName] = propMap

		// Check if this property should be required
		// Note: This is a simplified approach - in practice, you might want
		// to pass required fields as a separate parameter
		for _, req := range definition.Required {
			if req == propName {
				required = append(required, propName)
				break
			}
		}
	}

	return mcp.ToolInputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

func (m *McpClient) ParseToolDefinition(tool mcp.ToolInputSchema) map[string]jsonschema.Definition {
	definitions := make(map[string]jsonschema.Definition)

	// Handle the case where tool.Properties might be nil
	if tool.Properties == nil {
		return definitions
	}

	// Convert each property from the MCP tool schema to jsonschema.Definition
	for propName, propValue := range tool.Properties {
		definition := jsonschema.Definition{}

		// Handle different property value types
		if propMap, ok := propValue.(map[string]interface{}); ok {
			// Set type if present
			if typeVal, exists := propMap["type"]; exists {
				if typeStr, ok := typeVal.(string); ok {
					switch typeStr {
					case "string":
						definition.Type = jsonschema.String
					case "object":
						definition.Type = jsonschema.Object
					case "number":
						definition.Type = jsonschema.Number
					case "integer":
						definition.Type = jsonschema.Integer
					case "boolean":
						definition.Type = jsonschema.Boolean
					case "array":
						definition.Type = jsonschema.Array
					default:
						// Default to string for unknown types
						definition.Type = jsonschema.String
					}
				}
			}

			// Set description if present
			if descVal, exists := propMap["description"]; exists {
				if descStr, ok := descVal.(string); ok {
					definition.Description = descStr
				}
			}

			// Set enum values if present
			if enumVal, exists := propMap["enum"]; exists {
				if enumSlice, ok := enumVal.([]interface{}); ok {
					stringEnum := make([]string, len(enumSlice))
					for i, v := range enumSlice {
						if str, ok := v.(string); ok {
							stringEnum[i] = str
						}
					}
					definition.Enum = stringEnum
				}
			}

			// Note: Format field is not supported in jsonschema.Definition
			// If format handling is needed, it would need to be implemented differently
		}

		definitions[propName] = definition
	}

	return definitions
}


