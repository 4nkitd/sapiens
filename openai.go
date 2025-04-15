package sapiens

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/openai/openai-go"
)

// OpenAI implements the LLMInterface for OpenAI
type OpenAI struct {
	Client       openai.Client
	SystemPrompt SystemPrompt
	ModelName    string
	APIKey       string
	MaxTokens    int
	Temperature  float32
}

// NewOpenAI creates a new instance of OpenAI LLM
func NewOpenAI(apiKey string, modelName string) *OpenAI {
	return &OpenAI{
		APIKey:      apiKey,
		ModelName:   modelName,
		MaxTokens:   1024, // Default max tokens
		Temperature: 0.7,  // Default temperature
	}
}

// Initialize sets up the OpenAI client
func (o *OpenAI) Initialize() error {
	o.Client = openai.NewClient(
	// option.WithBaseURL("https://models.inference.ai.azure.com"), // defaults to https://api.openai.com
	// option.WithAPIKey(o.APIKey),                                 // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	return nil
}

func (o *OpenAI) GetModelName() string {
	return o.ModelName
}

func (o *OpenAI) GenerateEmbedding(ctx context.Context, model string, text string, embeddingType EmbeddingType) (Embedding, error) {

	responseOfAi, errEmbedding := o.Client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
	},
	)
	if errEmbedding != nil {
		return Embedding{}, errEmbedding
	}

	// Convert the response to Embedding
	if len(responseOfAi.Data) == 0 {
		return Embedding{}, fmt.Errorf("no embeddings returned from the API")
	}

	embedding := Embedding{
		Vector: responseOfAi.Data[0].Embedding,
		Text:   text,
		Type:   embeddingType,
	}

	return embedding, nil
}

// Complete generates a completion for the given prompt using default settings
func (o *OpenAI) Complete(ctx context.Context, prompt string) (string, error) {
	return o.CompleteWithOptions(ctx, prompt, nil)
}

// SetSystemPrompt sets the system prompt for the model
func (o *OpenAI) SetSystemPrompt(prompt SystemPrompt) {
	o.SystemPrompt = prompt
}

// CompleteWithOptions generates a completion with specific parameters
func (o *OpenAI) CompleteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("what is the capital of the United States?"),
		},
		Model: openai.ChatModelGPT4o,
	}

	if options != nil {
		// if temp, ok := options["temperature"].(float32); ok {
		// 	params.Temperature = openai.Float(float64(temp))
		// }
		// if maxTokens, ok := options["max_tokens"].(int); ok {
		// 	params.MaxTokens = openai.Int(int64(maxTokens))
		// }
	} else {
		// params.Temperature = openai.Float(float64(o.Temperature))
		// params.MaxTokens = openai.Int(int64(o.MaxTokens))
	}

	resp, err := o.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from the model")
	}

	return resp.Choices[0].Message.Content, nil
}

// ChatCompletion generates a response based on chat messages
func (o *OpenAI) ChatCompletion(ctx context.Context, messages []Message) (Response, error) {
	openaiMessages := []openai.ChatCompletionMessageParamUnion{}

	// Convert sapiens messages to OpenAI messages
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		default:
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))

		}
	}

	// Add system prompt if it exists
	if o.SystemPrompt.Content != "" {
		openaiMessages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(o.SystemPrompt.Content)}, openaiMessages...)
	}

	params := openai.ChatCompletionNewParams{
		Messages:    openaiMessages,
		Model:       openai.ChatModelGPT4o, // Assuming GPT-4o is the desired model
		Temperature: openai.Float(float64(o.Temperature)),
		MaxTokens:   openai.Int(int64(o.MaxTokens)),
	}

	resp, err := o.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return Response{}, fmt.Errorf("failed to generate chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("no choices returned from the model")
	}

	// Extract response content
	responseContent := resp.Choices[0].Message.Content

	// Create sapiens response
	response := Response{
		Content: responseContent,
		Raw:     resp,
	}

	return response, nil
}

// ChatCompletionWithTools generates a response with tools support
func (o *OpenAI) ChatCompletionWithTools(ctx context.Context, messages []Message, tools []Tool, options map[string]interface{}) (Response, error) {
	openaiMessages := []openai.ChatCompletionMessageParamUnion{}

	// Convert sapiens messages to OpenAI messages
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		default:
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		}
	}
	// Add system prompt if it exists
	if o.SystemPrompt.Content != "" {
		openaiMessages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(o.SystemPrompt.Content)}, openaiMessages...)
	}

	// Convert sapiens tools to OpenAI tools
	var openaiTools []openai.ChatCompletionToolParam
	for _, tool := range tools {
		openaiTool := openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
			},
		}

		// Convert sapiens input schema to OpenAI parameters
		if tool.InputSchema != nil {
			parameters := map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   tool.InputSchema.Required,
			}

			for key, prop := range tool.InputSchema.Properties {
				property := map[string]interface{}{
					"type":        prop.Type,
					"description": prop.Description,
				}

				if len(prop.Enum) > 0 {
					property["enum"] = prop.Enum
				}

				parameters["properties"].(map[string]interface{})[key] = property
			}

			openaiTool.Function.Parameters = parameters
		}

		openaiTools = append(openaiTools, openaiTool)
	}

	// Call OpenAI ChatCompletion API with tools
	params := openai.ChatCompletionNewParams{
		Model:       o.ModelName,
		Messages:    openaiMessages,
		Tools:       openaiTools,
		MaxTokens:   openai.Int(int64(o.MaxTokens)),
		Temperature: openai.Float(float64(o.Temperature)),
	}

	resp, err := o.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return Response{}, fmt.Errorf("failed to generate chat completion with tools: %v", err)
	}

	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("no choices returned from the model")
	}

	// Extract response content and tool calls
	responseContent := resp.Choices[0].Message.Content
	var toolCalls []ToolCall
	for _, toolCallResp := range resp.Choices[0].Message.ToolCalls {
		toolCalls = append(toolCalls, ToolCall{
			ID:       toolCallResp.ID,
			Name:     toolCallResp.Function.Name,
			InputMap: map[string]interface{}{}, // TODO: Populate input map from toolCallResp.Function.Arguments
		})
	}

	// Create sapiens response
	response := Response{
		Content:   responseContent,
		ToolCalls: toolCalls,
		Raw:       resp,
	}

	return response, nil
}

// ChatCompletionWithToolsAndHandlers generates a response with tools support and executes tool handlers
func (o *OpenAI) ChatCompletionWithToolsAndHandlers(ctx context.Context, messages []Message, tools []Tool,
	toolHandlers map[string]ToolHandler, options map[string]interface{}) (Response, error) {

	// Get initial response with tool calls
	response, err := o.ChatCompletionWithTools(ctx, messages, tools, options)
	if err != nil {
		return response, err
	}

	// If no tool calls or no handlers provided, return the response as is
	if len(response.ToolCalls) == 0 || len(toolHandlers) == 0 {
		return response, nil
	}

	// Process tool calls and collect results
	toolResults := make([]Message, 0, len(response.ToolCalls))
	for _, toolCall := range response.ToolCalls {
		// Check if we have a handler for this tool
		handler, exists := toolHandlers[toolCall.Name]
		if !exists {
			continue // Skip tools without handlers
		}

		// Execute the handler
		result, err := handler(toolCall.Name, toolCall.InputMap)
		if err != nil {
			// Add error as tool result
			toolResults = append(toolResults, Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolCall.Name,
				Content:    fmt.Sprintf("Error executing tool %s: %v", toolCall.Name, err),
			})
			continue
		}

		// Add successful result
		toolResults = append(toolResults, Message{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Name:       toolCall.Name,
			Content:    result,
		})
	}

	// If we have tool results, send them back to the model along with the original conversation
	if len(toolResults) > 0 {
		// Create a new conversation that includes the original messages, the assistant's response with tool calls,
		// and the tool results
		newMessages := make([]Message, 0, len(messages)+1+len(toolResults))

		// Add original messages
		newMessages = append(newMessages, messages...)

		// Add assistant's response with tool calls
		assistantMsg := Message{
			Role:    "assistant",
			Content: response.Content,
		}

		// If the response had tool calls, we need to include them
		if len(response.ToolCalls) > 0 {
			assistantMsg.ToolCalls = response.ToolCalls
			// Clear content as it will be represented by tool calls
			if assistantMsg.Content == "" {
				assistantMsg.Content = "I'll help you with that."
			}
		}
		newMessages = append(newMessages, assistantMsg)

		// Add tool results
		newMessages = append(newMessages, toolResults...)

		// Get final response
		finalResponse, err := o.ChatCompletion(ctx, newMessages)
		if err != nil {
			return response, fmt.Errorf("failed to get final response after tool execution: %v", err)
		}

		// Combine information from both responses
		response.Content = finalResponse.Content
		response.Raw = finalResponse.Raw
		response.ToolResults = toolResults

		return response, nil
	}

	return response, nil
}

// StructuredOutput generates a structured response based on a schema
func (o *OpenAI) StructuredOutput(ctx context.Context, messages []Message, schema Schema) (Response, error) {
	// No need to convert schema to JSON schema since we're not using it

	// Create a tool definition for structured output
	tools := []Tool{{
		Name:        "structured_output",
		Description: "Returns structured output based on the schema",
		InputSchema: &schema,
	}}

	// Create a tool handler (this is a dummy handler as OpenAI doesn't execute tools directly for structured output)
	toolHandlers := map[string]ToolHandler{
		"structured_output": func(toolName string, arguments map[string]interface{}) (string, error) {
			// This handler is not actually executed by OpenAI, but it's needed for the framework
			output, err := json.Marshal(arguments)
			return string(output), err
		},
	}

	// Call ChatCompletionWithToolsAndHandlers
	response, err := o.ChatCompletionWithToolsAndHandlers(ctx, messages, tools, toolHandlers, nil)
	if err != nil {
		return Response{}, err
	}

	// Extract structured data from the response
	var structuredData map[string]interface{}
	err = json.Unmarshal([]byte(response.Content), &structuredData)
	if err != nil {
		return Response{}, fmt.Errorf("failed to unmarshal structured data: %v", err)
	}

	response.Structured = structuredData

	return response, nil
}

// Generate implements LLMInterface.
func (o *OpenAI) Generate(ctx context.Context, request *Request) (*Response, error) {
	// Convert sapiens messages to OpenAI messages
	var openaiMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range request.Messages {
		switch msg.Role {
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		default:
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		}
	}

	// Add system prompt if it exists
	if o.SystemPrompt.Content != "" {
		openaiMessages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(o.SystemPrompt.Content)}, openaiMessages...)
	}

	// Call OpenAI ChatCompletion API
	params := openai.ChatCompletionNewParams{
		Model:       o.ModelName,
		Messages:    openaiMessages,
		MaxTokens:   openai.Int(int64(o.MaxTokens)),
		Temperature: openai.Float(float64(o.Temperature)),
	}

	resp, err := o.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from the model")
	}

	// Extract response content
	responseString := resp.Choices[0].Message.Content

	// Process response
	response := &Response{
		Content: responseString,
		Raw:     resp,
	}

	return response, nil
}

// MultimodalCompletion processes text and image inputs to generate a response
func (o *OpenAI) MultimodalCompletion(ctx context.Context, messages []Message) (Response, error) {

	return Response{}, nil
}
