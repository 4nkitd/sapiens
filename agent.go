package sapiens

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AgentFunc func(parameters map[string]string) string

type AgentTool struct {
	ToolDefinition openai.Tool
	ToolFunction   AgentFunc
}

type AToolCallResp struct {
	Name     string
	Id       string
	Response string
}

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

func NewAgent(ctx context.Context, llm *openai.Client, model string, systemPrompt string) *Agent {
	instance_of_agent := &Agent{
		Context:          ctx,
		Llm:              llm,
		Model:            model,
		SystemPrompt:     systemPrompt,
		maxToolCallDepth: 5, // Prevent infinite recursion
		currentDepth:     0,
	}

	return instance_of_agent
}

func (a *Agent) AddTool(name, description string, tool_parameters map[string]jsonschema.Definition, required_params []string, funx AgentFunc) error {
	tool_definition := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: tool_parameters,
				Required:   required_params,
			},
		},
	}

	agentTool := AgentTool{
		ToolDefinition: tool_definition,
		ToolFunction:   funx,
	}

	a.mu.Lock()
	a.Tools = append(a.Tools, agentTool)
	a.mu.Unlock()

	return nil
}

func (a *Agent) AddMCP(url string, customHeaders map[string]string) error {
	mcpClient, err := NewMcpClient(a.Context, url)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	// Get available tools from MCP server
	toolsResult, err := mcpClient.ListTools()
	if err != nil {
		return fmt.Errorf("failed to list MCP tools: %w", err)
	}

	a.mu.Lock()
	a.McpClient = mcpClient
	a.McpTools = toolsResult.Tools
	a.mu.Unlock()

	return nil
}

func (a *Agent) SetResponseSchema(name, description string, strict bool, defined_schema interface{}) *openai.ChatCompletionResponseFormat {
	schema, err := jsonschema.GenerateSchemaForType(defined_schema)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}

	msgSchema := &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Name:   name, // Fixed: use parameter instead of hardcoded value
			Schema: schema,
			Strict: strict,
		},
	}

	a.StructuredResponseSchema = msgSchema

	return msgSchema
}

func (a *Agent) ParseResponse(agent_response openai.ChatCompletionResponse, defined_schema interface{}) error {
	// Fixed: Add bounds checking
	if len(agent_response.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}

	return json.Unmarshal([]byte(agent_response.Choices[0].Message.Content), &defined_schema)
}

func (a *Agent) Ask(user_messages []openai.ChatCompletionMessage) (response openai.ChatCompletionResponse, err error) {
	system_message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.SystemPrompt,
		},
	}

	all_messages := append(system_message, user_messages...)

	a.mu.Lock()
	a.MessagesHistory = append(a.MessagesHistory, all_messages...)
	a.currentDepth = 0 // Reset depth for new conversation
	a.mu.Unlock()

	requestData := openai.ChatCompletionRequest{
		Model:    a.Model,
		Messages: a.MessagesHistory,
	}

	if a.StructuredResponseSchema != nil {
		requestData.ResponseFormat = a.StructuredResponseSchema
	}

	if len(a.Tools) > 0 || len(a.McpTools) > 0 {
		var openaiTools []openai.Tool

		a.mu.Lock()
		// Add regular tools
		for _, tool := range a.Tools {
			openaiTools = append(openaiTools, tool.ToolDefinition)
		}

		// Add MCP tools converted to OpenAI format
		for _, mcpTool := range a.McpTools {
			parsedProperties := a.McpClient.ParseToolDefinition(mcpTool.InputSchema)

			// Extract required fields from the MCP tool schema
			var requiredFields []string
			if mcpTool.InputSchema.Required != nil {
				for _, req := range mcpTool.InputSchema.Required {
					requiredFields = append(requiredFields, req)
				}
			}

			openaiTool := openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        mcpTool.Name,
					Description: mcpTool.Description,
					Parameters: jsonschema.Definition{
						Type:       jsonschema.Object,
						Properties: parsedProperties,
						Required:   requiredFields,
					},
				},
			}
			openaiTools = append(openaiTools, openaiTool)
		}
		a.mu.Unlock()

		requestData.Tools = openaiTools
	}

	a.Request = requestData

	fmt.Printf("%+v", a.Request)

	return a.AskAi(a.Context)
}

func (a *Agent) AskAi(ctx context.Context) (openai.ChatCompletionResponse, error) {
	a.mu.Lock()
	a.Request.Messages = a.MessagesHistory
	a.mu.Unlock()

	responseStr, responseErr := a.Llm.CreateChatCompletion(
		ctx, // Fixed: Use the passed context parameter
		a.Request,
	)

	if responseErr != nil {
		return responseStr, responseErr
	}

	// Process tool calls if any and return the final response
	finalResponse, err := a.ToolCalls(responseStr)
	if err != nil {
		return responseStr, fmt.Errorf("tool call processing error: %w", err)
	}

	// Return the final response if tools were called, otherwise return original response
	if finalResponse != nil {
		return *finalResponse, nil
	}

	return responseStr, responseErr
}

func (a *Agent) ToolCalls(response openai.ChatCompletionResponse) (*openai.ChatCompletionResponse, error) {
	// Fixed: Add recursion depth check to prevent infinite loops
	if a.currentDepth >= a.maxToolCallDepth {
		return nil, fmt.Errorf("maximum tool call depth (%d) exceeded", a.maxToolCallDepth)
	}

	var toolResponses []AToolCallResp
	var totalToolExecCount int = 0

	// Check if response has function calls
	for _, choice := range response.Choices {
		if choice.Message.ToolCalls != nil && len(choice.Message.ToolCalls) > 0 {
			// Don't add assistant message with tool calls for Gemini compatibility

			for _, toolCall := range choice.Message.ToolCalls {
				// First try to find regular tool
				toolInst, toolInsErr := a.GetToolByName(toolCall.Function.Name)
				if toolInsErr == nil {
					// Regular tool found
					var parsedParams map[string]string
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parsedParams); err != nil {
						return nil, fmt.Errorf("failed to parse tool arguments for '%s': %w", toolCall.Function.Name, err)
					}

					toolResponse := toolInst.ToolFunction(parsedParams)

					toolResponses = append(toolResponses, AToolCallResp{
						Response: toolResponse,
						Id:       toolCall.ID,
						Name:     toolCall.Function.Name,
					})
				} else {
					// Try MCP tool
					mcpTool, mcpErr := a.GetMcpToolByName(toolCall.Function.Name)
					if mcpErr != nil {
						return nil, fmt.Errorf("tool '%s' not found in regular or MCP tools: %w", toolCall.Function.Name, mcpErr)
					}

					// Parse arguments as generic map for MCP
					var parsedArgs map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parsedArgs); err != nil {
						return nil, fmt.Errorf("failed to parse MCP tool arguments for '%s': %w", toolCall.Function.Name, err)
					}

					// Call MCP tool
					mcpResult, mcpCallErr := a.McpClient.CallTool(mcp.CallToolParams{
						Name:      mcpTool.Name,
						Arguments: parsedArgs,
					})

					if mcpCallErr != nil {
						return nil, fmt.Errorf("MCP tool call failed for '%s': %w", toolCall.Function.Name, mcpCallErr)
					}

					// Convert MCP result to string
					var toolResponse string
					if len(mcpResult.Content) > 0 {
						toolResponse = fmt.Sprintf("%v", mcpResult.Content[0])
					} else {
						toolResponse = "MCP tool executed successfully"
					}

					toolResponses = append(toolResponses, AToolCallResp{
						Response: toolResponse,
						Id:       toolCall.ID,
						Name:     toolCall.Function.Name,
					})
				}

				totalToolExecCount++
			}
		}
	}

	// Fixed: Add tool responses using user message format for Gemini compatibility
	if len(toolResponses) > 0 {
		a.mu.Lock()
		for _, agentToolResp := range toolResponses {
			// Use user message format instead of tool message for Gemini compatibility
			toolMessage := NewMessages().UserMessage(
				fmt.Sprintf("Tool '%s' returned: %s", agentToolResp.Name, agentToolResp.Response),
			)
			a.MessagesHistory = append(a.MessagesHistory, toolMessage)
		}
		a.currentDepth++ // Increment depth before recursive call
		a.mu.Unlock()

		// Fixed: Recursive call with proper termination condition and return final response
		if totalToolExecCount > 0 {
			finalResponse, err := a.AskAi(a.Context)
			if err != nil {
				return nil, err
			}
			return &finalResponse, nil
		}
	}

	// No tool calls found, return nil to indicate original response should be used
	return nil, nil
}

func (a *Agent) GetToolByName(name string) (AgentTool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, tool := range a.Tools {
		if tool.ToolDefinition.Function.Name == name {
			return tool, nil
		}
	}

	return AgentTool{}, fmt.Errorf("tool not found")
}

func (a *Agent) GetMcpToolByName(name string) (mcp.Tool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, tool := range a.McpTools {
		if tool.Name == name {
			return tool, nil
		}
	}

	return mcp.Tool{}, fmt.Errorf("MCP tool not found")
}
