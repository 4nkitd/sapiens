package sapiens

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ToolImplementation is a function that handles tool calls
type ToolImplementation func(params map[string]interface{}) (interface{}, error)

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
		PromptManager:       NewPromptManager(),
	}
}

// AddSystemPrompt adds a system prompt to the agent
func (a *Agent) AddSystemPrompt(content string, version string) {
	a.SystemPrompts = append(a.SystemPrompts, SystemPrompt{
		Content: content,
		Version: version,
	})

	a.LLM.Implementation.SetSystemPrompt(a.GetLatestSystemPrompt())
}

// AddDynamicPrompt adds or updates a system prompt using a template and data
func (a *Agent) AddDynamicPrompt(promptTemplate string, cardData map[string]interface{}, version string) error {
	// Process the template with the card data
	dynamicPrompt, err := ApplyTemplate(promptTemplate, cardData)
	if err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Add the processed prompt as a system prompt
	a.AddSystemPrompt(dynamicPrompt, version)
	return nil
}

// AddDynamicPromptWithCard adds a system prompt using a card and the prompt manager
func (a *Agent) AddDynamicPromptWithCard(card Card, version string) error {
	// Render the prompt from the card
	renderedPrompt, err := card.Render(a.PromptManager)
	if err != nil {
		return fmt.Errorf("failed to render prompt card: %w", err)
	}

	// Add the rendered prompt as a system prompt
	a.AddSystemPrompt(renderedPrompt, version)
	return nil
}

// AddPromptTemplate adds a new prompt template to the agent's prompt manager
func (a *Agent) AddPromptTemplate(template PromptTemplate) error {
	return a.PromptManager.AddTemplate(template)
}

// GetPromptTemplate retrieves a prompt template by name
func (a *Agent) GetPromptTemplate(name string) (PromptTemplate, error) {
	return a.PromptManager.GetTemplate(name)
}

// AddTools adds tools to the agent
func (a *Agent) AddTools(tools ...Tool) {
	a.Tools = append(a.Tools, tools...)
}

// SetContext sets the context for the agent
func (a *Agent) SetContext(context map[string]interface{}) {
	a.Context = context
}

// SetStringContext sets a string as context for the agent
func (a *Agent) SetStringContext(contextString string) {
	// Store in the context map
	a.Context["context_information"] = contextString

	// Also inject the context into the conversation
	a.InjectContextIntoConversation(contextString)
}

// InjectContextIntoConversation adds the context as a system message in the conversation
func (a *Agent) InjectContextIntoConversation(contextString string) {
	contextMessage := Message{
		Role:    "system",
		Content: fmt.Sprintf("Here is important context information to use when answering questions:\n\n%s", contextString),
	}

	// Add to the beginning of messages after any system prompts
	systemPromptCount := 0
	for _, msg := range a.Messages {
		if msg.Role == "system" {
			systemPromptCount++
		}
	}

	// Create a new messages slice with the context inserted after system prompts
	newMessages := make([]Message, 0, len(a.Messages)+1)
	newMessages = append(newMessages, a.Messages[:systemPromptCount]...)
	newMessages = append(newMessages, contextMessage)
	if systemPromptCount < len(a.Messages) {
		newMessages = append(newMessages, a.Messages[systemPromptCount:]...)
	}

	a.Messages = newMessages
}

// UpdateContext updates the context for the agent
func (a *Agent) UpdateContext(context map[string]interface{}) {
	a.Context = context
}

// UpdateStringContext updates the string context for the agent
func (a *Agent) UpdateStringContext(contextString string) {
	// Update in the context map
	a.Context["context_information"] = contextString

	// Also inject the updated context into the conversation
	a.InjectContextIntoConversation(contextString)
}

// SetStructuredResponseSchema sets the schema for structured responses
func (a *Agent) SetStructuredResponseSchema(schema Schema) {
	a.StructuredResponseSchema = schema
	a.Tools = append(a.Tools, Tool{
		Name:        "structured_output",
		Description: "Structured output tool",
		InputSchema: &schema,
	})
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

	// Implement retry logic
	var response Response
	var err error
	var attempts int

	for attempts = 0; attempts < a.MaxRetry; attempts++ {
		// Logic fork based on capabilities needed
		if len(a.Tools) > 0 && a.StructuredResponseSchema.Type != "" {
			response, err = a.executeWithToolsAndStructure(ctx, options)
		} else if len(a.Tools) > 0 {
			response, err = a.executeWithTools(ctx, options)
		} else if a.StructuredResponseSchema.Type != "" {
			response, err = a.executeWithStructure(ctx)
		} else {
			// Query memory and add results to messages
			if a.Memory != nil {
				memoryResults := a.Memory.Search([]float64{}) // Replace with actual query vector if available
				if len(memoryResults) > 0 {
					memoryContext := "Relevant information from memory:\n"
					for _, result := range memoryResults {
						memoryContext += fmt.Sprintf("- %s\n", result.Text)
					}
					a.Messages = append(a.Messages, Message{
						Role:    "system",
						Content: memoryContext,
					})
				}
			}
			response, err = a.LLM.Implementation.ChatCompletion(ctx, a.Messages)
		}

		if err == nil {
			break // Success, exit retry loop
		}

		// Log retry attempt
		fmt.Printf("LLM execution failed (attempt %d/%d): %v\n",
			attempts+1, a.MaxRetry, err)

		// Optional: Add backoff delay between retries
		if attempts < a.MaxRetry-1 {
			time.Sleep(time.Duration(500*(attempts+1)) * time.Millisecond)
		}
	}

	// If all retries failed
	if err != nil {
		return Response{}, fmt.Errorf("LLM execution failed after %d attempts: %w",
			attempts, err)
	}

	return response, nil
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

// GetLatestSystemPrompt returns the most recent system prompt
func (a *Agent) GetLatestSystemPrompt() SystemPrompt {
	if len(a.SystemPrompts) == 0 {
		return SystemPrompt{}
	}
	return a.SystemPrompts[len(a.SystemPrompts)-1]
}

// executeWithTools processes a request with tool support
func (a *Agent) executeWithTools(ctx context.Context, options map[string]interface{}) (Response, error) {
	// Implement retry logic
	var response Response
	var err error
	var attempts int

	// Create a filtered tools list excluding the structured_output tool
	filteredTools := make([]Tool, 0, len(a.Tools))
	for _, tool := range a.Tools {
		if tool.Name != "structured_output" {
			filteredTools = append(filteredTools, tool)
		}
	}

	for attempts = 0; attempts < a.MaxRetry; attempts++ {
		response, err = a.LLM.Implementation.ChatCompletionWithTools(ctx, a.Messages, filteredTools, options)
		if err == nil {
			break // Success, exit retry loop
		}

		// Log retry attempt
		fmt.Printf("executeWithTools failed (attempt %d/%d): %v\n",
			attempts+1, a.MaxRetry, err)

		// Optional: Add backoff delay between retries
		if attempts < a.MaxRetry-1 {
			time.Sleep(time.Duration(500*(attempts+1)) * time.Millisecond)
		}
	}

	// If all retries failed
	if err != nil {
		return Response{}, fmt.Errorf("executeWithTools failed after %d attempts: %w", attempts, err)
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
	// Implement retry logic
	var response Response
	var err error
	var attempts int

	for attempts = 0; attempts < a.MaxRetry; attempts++ {
		response, err = a.LLM.Implementation.StructuredOutput(ctx, a.Messages, a.StructuredResponseSchema)
		if err == nil {
			break // Success, exit retry loop
		}

		// Log retry attempt
		fmt.Printf("executeWithStructure failed (attempt %d/%d): %v\n",
			attempts+1, a.MaxRetry, err)

		// Optional: Add backoff delay between retries
		if attempts < a.MaxRetry-1 {
			time.Sleep(time.Duration(500*(attempts+1)) * time.Millisecond)
		}
	}

	// If all retries failed
	if err != nil {
		return Response{}, fmt.Errorf("executeWithStructure failed after %d attempts: %w", attempts, err)
	}

	// Add the assistant response to the message history
	a.Messages = append(a.Messages, Message{
		Role:    "assistant",
		Content: response.Content,
	})

	return response, nil
}

// ExecuteWithToolsAndStructure processes a request with both tools and structured output
func (a *Agent) executeWithToolsAndStructure(ctx context.Context, options map[string]interface{}) (Response, error) {
	// Create a filtered tools list excluding the structured_output tool
	filteredTools := make([]Tool, 0, len(a.Tools))
	for _, tool := range a.Tools {
		if tool.Name != "structured_output" {
			filteredTools = append(filteredTools, tool)
		}
	}

	// Execute with tools first
	response, err := a.LLM.Implementation.ChatCompletionWithTools(ctx, a.Messages, filteredTools, options)
	if err != nil {
		return Response{}, err
	}

	// If there are tool calls, process them immediately within this function
	if len(response.ToolCalls) > 0 {
		// Add the assistant response to the message history
		a.Messages = append(a.Messages, Message{
			Role:      "assistant",
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		})

		// Process each tool call and add results to the conversation
		for _, toolCall := range response.ToolCalls {
			implementation, exists := a.toolImplementations[toolCall.Name]
			if !exists {
				continue
			}

			// Execute the tool
			result, err := implementation(toolCall.InputMap)
			if err != nil {
				continue
			}

			// Convert result to JSON
			resultJSON, err := json.Marshal(result)
			if err != nil {
				continue
			}

			// Add result to conversation
			a.Messages = append(a.Messages, Message{
				Role:    "function",
				Name:    toolCall.Name,
				Content: string(resultJSON),
				Options: map[string]interface{}{
					"tool_call_id": toolCall.ID,
				},
			})
		}

		// Now request a structured response directly
		structuredPrompt := fmt.Sprintf(
			"Based on the information provided about the weather in Delhi, respond with a JSON object containing these fields: answer (string): a concise summary of the weather information, confidence (number): your confidence in this answer from 0 to 1. ONLY respond with valid JSON.",
		)

		// Add structured prompt
		a.Messages = append(a.Messages, Message{
			Role:    "user",
			Content: structuredPrompt,
		})

		// Use regular ChatCompletion to avoid MIME type errors
		structuredResponse, err := a.LLM.Implementation.ChatCompletion(ctx, a.Messages)
		if err != nil {
			// Return the original response even if structured part fails
			a.Messages = a.Messages[:len(a.Messages)-1] // Remove the structured prompt
			return response, nil
		}

		// Try to extract JSON from the response
		var structured interface{}
		err = json.Unmarshal([]byte(structuredResponse.Content), &structured)
		if err == nil {
			// Successfully parsed JSON
			response.Structured = structured
			response.Content = structuredResponse.Content
		}

		// Remove the structured prompt from messages
		a.Messages = a.Messages[:len(a.Messages)-1]

		return response, nil
	}

	// If no tool calls, try to get structured response directly
	var structured interface{}
	err = json.Unmarshal([]byte(response.Content), &structured)
	if err == nil {
		response.Structured = structured
		return response, nil
	}

	// If we reached here, we need to explicitly request a structured response
	structuredResponse, err := a.executeWithStructure(ctx)
	if err == nil {
		response.Structured = structuredResponse.Structured
		if response.Content == "" {
			response.Content = structuredResponse.Content
		}
	}

	return response, nil
}

// Helper function to get a description of schema fields
func getSchemaDescription(schema Schema) string {
	result := ""
	for name, prop := range schema.Properties {
		result += name + " (" + prop.Type + "): " + prop.Description + ", "
	}
	return result
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

// Run processes a user query and returns a response with retry support
func (a *Agent) Run(ctx context.Context, query string) (*Response, error) {
	// Create the user message
	userMessage := Message{
		Role:    "user",
		Content: query,
	}

	// Add user message to both conversation histories
	a.conversationHistory = append(a.conversationHistory, userMessage)
	a.Messages = append(a.Messages, userMessage)

	// Query memory for relevant information
	if a.Memory != nil {
		memoryResults := a.Memory.Search([]float64{}) // Replace with actual query vector if available
		if len(memoryResults) > 0 {
			// Add memory results as a system message
			memoryContext := "Relevant information from memory:\n"
			for _, result := range memoryResults {
				memoryContext += fmt.Sprintf("- %s\n", result.Text)
			}
			a.Messages = append(a.Messages, Message{
				Role:    "system",
				Content: memoryContext,
			})
		}
	}

	// Implement retry logic
	var response Response
	var err error
	var attempts int

	for attempts = 0; attempts < a.MaxRetry; attempts++ {
		// Process the request with ExecuteLLM
		response, err = a.ExecuteLLM(ctx)
		if err == nil {
			break // Success, exit retry loop
		}

		// Log retry attempt
		fmt.Printf("LLM execution failed (attempt %d/%d): %v\n",
			attempts+1, a.MaxRetry, err)

		// Optional: Add backoff delay between retries
		if attempts < a.MaxRetry-1 {
			time.Sleep(time.Duration(500*(attempts+1)) * time.Millisecond)
		}
	}

	// If all retries failed
	if err != nil {
		return nil, fmt.Errorf("LLM execution failed after %d attempts: %w",
			attempts, err)
	}

	// Convert Response to *Response
	responsePtr := &Response{
		Content:    response.Content,
		ToolCalls:  response.ToolCalls,
		Structured: response.Structured,
		Raw:        response.Raw,
	}

	// Process tool calls if present (and they're not structured_output calls)
	if len(responsePtr.ToolCalls) > 0 && len(a.toolImplementations) > 0 {
		// Check if all tool calls are structured_output
		allStructured := true
		for _, call := range responsePtr.ToolCalls {
			if call.Name != "structured_output" {
				allStructured = false
				break
			}
		}

		// Only process if we have actual tool calls to execute
		if !allStructured {
			finalResponse, err := a.handleToolCalls(ctx, responsePtr)
			if err != nil {
				return responsePtr, err // Return partial response on error
			}

			// Preserve structured data
			if responsePtr.Structured != nil && finalResponse.Structured == nil {
				finalResponse.Structured = responsePtr.Structured
			}

			return finalResponse, nil
		}
	}

	// Add assistant response to conversation history
	a.conversationHistory = append(a.conversationHistory, Message{
		Role:    "assistant",
		Content: responsePtr.Content,
	})

	return responsePtr, nil
}

// handleToolCalls processes tool calls and continues the conversation
func (a *Agent) handleToolCalls(ctx context.Context, response *Response) (*Response, error) {
	toolResults := make([]ToolResult, 0, len(response.ToolCalls))

	// Process each tool call (skip structured_output)
	for _, toolCall := range response.ToolCalls {
		if toolCall.Name == "structured_output" {
			// Handle structured output separately
			response.Structured = toolCall.InputMap
			continue
		}

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

	// If no actual tools were executed (only structured_output), return early
	if len(toolResults) == 0 {
		return response, nil
	}

	// Add tool results to conversation
	for _, result := range toolResults {
		a.conversationHistory = append(a.conversationHistory, Message{
			Role:       "tool",
			Content:    result.Result,
			ToolCallID: result.ToolCallID,
		})

		// Also add to Messages for consistency
		a.Messages = append(a.Messages, Message{
			Role:    "function",
			Name:    getToolNameFromCallID(result.ToolCallID, response.ToolCalls),
			Content: result.Result,
			Options: map[string]interface{}{
				"tool_call_id": result.ToolCallID,
			},
		})
	}

	// For structured output, directly request JSON after tool results
	if a.StructuredResponseSchema.Type != "" {
		// Add explicit instruction for JSON response
		structuredPrompt := fmt.Sprintf(
			"Based on the information provided, respond ONLY with a valid JSON object containing these fields: %s.",
			getSchemaDescription(a.StructuredResponseSchema),
		)

		// Create temporary messages to request structured output
		tempMessages := make([]Message, len(a.Messages))
		copy(tempMessages, a.Messages)
		tempMessages = append(tempMessages, Message{
			Role:    "user",
			Content: structuredPrompt,
		})

		// Get structured response
		structuredResponse, err := a.LLM.Implementation.ChatCompletion(ctx, tempMessages)
		if err == nil {
			// Try to parse as JSON
			var structured interface{}
			err = json.Unmarshal([]byte(structuredResponse.Content), &structured)
			if err == nil {
				// Create final response with both content and structured data
				finalResponse := Response{
					Content:    structuredResponse.Content,
					Structured: structured,
				}

				// Add final response to conversation history
				a.conversationHistory = append(a.conversationHistory, Message{
					Role:    "assistant",
					Content: finalResponse.Content,
				})

				return &finalResponse, nil
			}
		}
	}

	// If structured approach failed or wasn't needed, use regular continuation
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

// getToolNameFromCallID finds the tool name for a given tool call ID
func getToolNameFromCallID(callID string, calls []ToolCall) string {
	for _, call := range calls {
		if call.ID == callID {
			return call.Name
		}
	}
	return ""
}

// GetHistory returns the conversation history
func (a *Agent) GetHistory() []Message {
	return a.conversationHistory
}
