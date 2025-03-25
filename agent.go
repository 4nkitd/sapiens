package sapiens

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ToolImplementation is a function that handles tool calls
type ToolImplementation func(params map[string]interface{}) (interface{}, error)

// Agent represents an AI agent that can process queries and use tools

// NewAgent creates a new agent instance
func NewAgent(name string, llmImplementation LLMInterface, apiKey, model, provider string) *Agent {
	return &Agent{
		Name: name,
		LLM: &LLM{
			Implementation: llmImplementation,
			ApiKey:         apiKey,
			Model:          model,
			Provider:       provider,
		},
		SystemPrompts:       []SystemPrompt{},
		Tools:               []Tool{},
		toolImplementations: make(map[string]ToolImplementation),
		conversationHistory: []Message{},
		MaxRetry:            3,
		Context:             make(map[string]interface{}),
		MetaData:            make(map[string]interface{}),
	}
}

// AddSystemPrompt adds a system prompt to the agent
func (a *Agent) AddSystemPrompt(content string, version string) {
	a.SystemPrompts = append(a.SystemPrompts, SystemPrompt{
		Content: content,
		Version: version,
	})

	a.Messages = append(a.Messages, Message{
		Role:    "system",
		Content: content,
	})
}

// AddTools adds tools to the agent
func (a *Agent) AddTools(tools ...Tool) {
	a.Tools = append(a.Tools, tools...)
}

// SetStructuredResponseSchema sets the schema for structured responses
func (a *Agent) SetStructuredResponseSchema(schema Schema) {
	a.StructuredResponseSchema = schema
}

// RegisterToolImplementation registers a tool implementation function
func (a *Agent) RegisterToolImplementation(toolName string, implementation ToolImplementation) {
	a.toolImplementations[toolName] = implementation
}

// ExecuteLLM processes the current agent state through the LLM and returns a response.
func (a *Agent) ExecuteLLM(ctx context.Context) (Response, error) {
	if a.LLM == nil || a.LLM.Implementation == nil {
		return Response{}, errors.New("LLM implementation not set")
	}

	// Create options from agent context
	options := a.getOptions()

	// Logic fork based on capabilities needed
	if len(a.Tools) > 0 {
		return a.executeWithTools(ctx, options)
	} else if a.StructuredResponseSchema.Type != "" {
		return a.executeWithStructure(ctx)
	} else {
		return a.LLM.Implementation.ChatCompletion(ctx, a.Messages)
	}
}

// getOptions extracts options from the agent context
func (a *Agent) getOptions() map[string]interface{} {
	options := make(map[string]interface{})
	for _, key := range []string{"temperature", "max_tokens", "top_p", "frequency_penalty", "presence_penalty"} {
		if val, ok := a.Context[key]; ok {
			options[key] = val
		}
	}
	return options
}

// executeWithTools processes a request with tool support
func (a *Agent) executeWithTools(ctx context.Context, options map[string]interface{}) (Response, error) {
	response, err := a.LLM.Implementation.ChatCompletionWithTools(ctx, a.Messages, a.Tools, options)
	if err != nil {
		return Response{}, err
	}

	// Add the assistant response to the message history
	a.Messages = append(a.Messages, Message{
		Role:      "assistant",
		Content:   response.Content,
		ToolCalls: response.ToolCalls,
	})

	return response, nil
}

// executeWithStructure processes a request with structured output
func (a *Agent) executeWithStructure(ctx context.Context) (Response, error) {
	response, err := a.LLM.Implementation.StructuredOutput(ctx, a.Messages, a.StructuredResponseSchema)
	if err != nil {
		return Response{}, err
	}

	// Add the assistant response to the message history
	a.Messages = append(a.Messages, Message{
		Role:    "assistant",
		Content: response.Content,
	})

	return response, nil
}

// executeWithToolsAndStructure processes a request with both tools and structured output
// This is more complex and might require custom handling depending on the LLM
func (a *Agent) ExecuteWithToolsAndStructure(ctx context.Context) (Response, error) {
	// First, execute with tools
	toolResponse, err := a.executeWithTools(ctx, a.getOptions())
	if err != nil {
		return Response{}, err
	}
	fmt.Println(toolResponse)
	// If there are tool calls, handle them first
	if len(toolResponse.ToolCalls) > 0 {
		return toolResponse, nil
	}

	// Otherwise, ensure the response is structured
	var structured interface{}
	err = json.Unmarshal([]byte(toolResponse.Content), &structured)
	if err != nil {
		// If not structured, try again with explicit structure request
		return a.executeWithStructure(ctx)
	}

	toolResponse.Structured = structured
	return toolResponse, nil
}

// HandleToolResponse processes tool outputs and continues the conversation
func (a *Agent) HandleToolResponse(ctx context.Context, toolCallID string, toolName string, result string) (Response, error) {
	// Add the function response to the conversation
	a.Messages = append(a.Messages, Message{
		Role:    "function",
		Name:    toolName,
		Content: result,
		Options: map[string]interface{}{
			"tool_call_id": toolCallID,
		},
	})

	// Continue the conversation
	return a.ExecuteLLM(ctx)
}

// Run processes a user query and returns a response at the end
func (a *Agent) Run(ctx context.Context, query string) (*Response, error) {
	// Create the user message
	userMessage := Message{
		Role:    "user",
		Content: query,
	}

	// Add user message to both conversation histories
	a.conversationHistory = append(a.conversationHistory, userMessage)
	a.Messages = append(a.Messages, userMessage)

	// Process the request with ExecuteLLM
	response, err := a.ExecuteLLM(ctx)
	if err != nil {
		return nil, fmt.Errorf("LLM execution failed: %w", err)
	}

	// Convert Response to *Response
	responsePtr := &Response{
		Content:    response.Content,
		ToolCalls:  response.ToolCalls,
		Structured: response.Structured,
	}

	// Process tool calls if present
	if len(responsePtr.ToolCalls) > 0 {

		fmt.Printf("Tool calls: %v\n", responsePtr)

		return a.handleToolCalls(ctx, responsePtr)
	}

	// Add assistant response to conversation history
	a.conversationHistory = append(a.conversationHistory, Message{
		Role:    "assistant",
		Content: responsePtr.Content,
	})

	return responsePtr, nil
}

// Helper function to extract tool names for debugging
func toolNames(tools []Tool) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}

// handleToolCalls processes tool calls and continues the conversation
func (a *Agent) handleToolCalls(ctx context.Context, response *Response) (*Response, error) {
	toolResults := make([]ToolResult, 0, len(response.ToolCalls))

	// Process each tool call
	for _, toolCall := range response.ToolCalls {
		result, err := a.executeTool(toolCall)
		if err != nil {
			return nil, err
		}
		toolResults = append(toolResults, result)
	}

	// Update conversation history with the assistant's tool calls
	a.conversationHistory = append(a.conversationHistory, Message{
		Role:      "assistant",
		Content:   response.Content,
		ToolCalls: response.ToolCalls,
	})

	// Add tool results to conversation history
	for _, result := range toolResults {
		a.conversationHistory = append(a.conversationHistory, Message{
			Role:       "tool",
			Content:    result.Result,
			ToolCallID: result.ToolCallID,
		})

		// Also add to Messages for consistency
		a.Messages = append(a.Messages, Message{
			Role:    "function", // LLM interfaces typically use "function" role
			Name:    getTooNameFromCallID(result.ToolCallID, response.ToolCalls),
			Content: result.Result,
			Options: map[string]interface{}{
				"tool_call_id": result.ToolCallID,
			},
		})
	}

	// Process the final response after tool calls
	finalResponse, err := a.ExecuteLLM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %w", err)
	}

	// Add final response to conversation history
	a.conversationHistory = append(a.conversationHistory, Message{
		Role:    "assistant",
		Content: finalResponse.Content,
	})

	// Return as pointer to match function signature
	return &finalResponse, nil
}

// executeTool runs a single tool call and returns the result
func (a *Agent) executeTool(toolCall ToolCall) (ToolResult, error) {
	// Get the tool implementation
	implementation, exists := a.toolImplementations[toolCall.Name]
	if !exists {
		return ToolResult{}, fmt.Errorf("no implementation found for tool: %s", toolCall.Name)
	}

	// Execute the tool
	result, err := implementation(toolCall.InputMap)
	if err != nil {
		return ToolResult{}, fmt.Errorf("tool execution failed: %w", err)
	}

	// Convert result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to marshal tool result: %w", err)
	}

	return ToolResult{
		ToolCallID: toolCall.ID,
		Result:     string(resultJSON),
	}, nil
}

// getTooNameFromCallID finds the tool name for a given tool call ID
func getTooNameFromCallID(callID string, calls []ToolCall) string {
	for _, call := range calls {
		if call.ID == callID {
			return call.Name
		}
	}
	return ""
}
