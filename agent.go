package sapiens

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AgentFunc func(parameters map[string]string) string

type AgentTool struct {
	ToolDefinition openai.Tool
	ToolFunction   AgentFunc
}

type Agent struct {
	MessagesHistory          []openai.ChatCompletionMessage
	Context                  context.Context
	Llm                      *openai.Client
	Model                    string
	SystemPrompt             string
	StructuredResponseSchema *openai.ChatCompletionResponseFormat
	Tools                    []AgentTool
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

	if len(a.Tools) > 0 {
		var openaiTools []openai.Tool

		a.mu.Lock()
		for _, tool := range a.Tools {
			// Fixed: Use the tool definition directly instead of reconstructing
			openaiTools = append(openaiTools, tool.ToolDefinition)
		}
		a.mu.Unlock()

		requestData.Tools = openaiTools
	}

	a.Request = requestData

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

type AToolCallResp struct {
	Name     string
	Id       string
	Response string
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
				toolInst, toolInsErr := a.GetToolByName(toolCall.Function.Name)
				if toolInsErr != nil {
					// Fixed: Handle tool not found error properly
					return nil, fmt.Errorf("tool '%s' not found: %w", toolCall.Function.Name, toolInsErr)
				}

				var parsedParams map[string]string
				// Fixed: Handle JSON unmarshal error
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parsedParams); err != nil {
					return nil, fmt.Errorf("failed to parse tool arguments for '%s': %w", toolCall.Function.Name, err)
				}

				toolResponse := toolInst.ToolFunction(parsedParams)

				toolResponses = append(toolResponses, AToolCallResp{
					Response: toolResponse,
					Id:       toolCall.ID,
					Name:     toolCall.Function.Name,
				})

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
