package sapiens

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ToolHandler is a function that handles tool calls from the LLM
type ToolHandler func(toolName string, arguments map[string]interface{}) (string, error)

// LLMInterface defines the contract for language model implementations.
type LLMInterface interface {
	// Initialize sets up the language model with provided options
	Initialize() error

	Generate(ctx context.Context, request *Request) (*Response, error)

	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)

	// CompleteWithOptions generates a completion with specific parameters
	CompleteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error)

	// ChatCompletion generates a response based on chat messages
	ChatCompletion(ctx context.Context, messages []Message) (Response, error)

	// ChatCompletionWithTools generates a response with tools support
	ChatCompletionWithTools(ctx context.Context, messages []Message, tools []Tool, options map[string]interface{}) (Response, error)

	// ChatCompletionWithToolsAndHandlers generates a response with tools support and executes tool handlers
	ChatCompletionWithToolsAndHandlers(ctx context.Context, messages []Message, tools []Tool,
		toolHandlers map[string]ToolHandler, options map[string]interface{}) (Response, error)

	// StructuredOutput generates a structured response based on a schema
	StructuredOutput(ctx context.Context, messages []Message, schema Schema) (Response, error)
}

// GoogleGenAI implements the LLMInterface for Google's Generative AI
type GoogleGenAI struct {
	Client      *genai.Client
	Model       *genai.GenerativeModel
	APIKey      string
	ModelName   string
	MaxTokens   int32
	Temperature float32
}

// Generate implements LLMInterface.
func (g *GoogleGenAI) Generate(ctx context.Context, request *Request) (*Response, error) {
	if g.Client == nil || g.Model == nil {
		return nil, fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	// Convert sapiens messages to genai chat messages
	var inputParts []genai.Part
	for _, msg := range request.Messages {
		inputParts = append(inputParts, genai.Text(msg.Content))
	}

	// Generate content
	resp, err := g.Model.GenerateContent(ctx, inputParts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from the model")
	}

	responseString := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseString += string(text)
		}
	}

	// Process response
	response := &Response{
		Content: responseString,
		Raw:     resp,
	}

	return response, nil
}

// NewGoogleGenAI creates a new instance of Google GenAI LLM
func NewGoogleGenAI(apiKey string, modelName string) *GoogleGenAI {
	return &GoogleGenAI{
		APIKey:      apiKey,
		ModelName:   modelName,
		MaxTokens:   1024, // Default max tokens
		Temperature: 0.7,  // Default temperature
	}
}

// Initialize sets up the Google GenAI client and model
func (g *GoogleGenAI) Initialize() error {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.APIKey))
	if err != nil {
		return fmt.Errorf("failed to create genai client: %v", err)
	}

	g.Client = client
	g.Model = client.GenerativeModel(g.ModelName)
	return nil
}

// Complete generates a completion for the given prompt using default settings
func (g *GoogleGenAI) Complete(ctx context.Context, prompt string) (string, error) {
	return g.CompleteWithOptions(ctx, prompt, nil)
}

// CompleteWithOptions generates a completion with specific parameters
func (g *GoogleGenAI) CompleteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	if g.Client == nil || g.Model == nil {
		return "", fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	// Create a prompt for the model
	genPrompt := []genai.Part{
		genai.Text(prompt),
	}

	// Setup generation config with options
	if options != nil {
		if temp, ok := options["temperature"].(float32); ok {
			g.Model.SetTemperature(temp)
		}
		if maxTokens, ok := options["max_tokens"].(int32); ok {
			g.Model.SetMaxOutputTokens(maxTokens)
		}
	} else {
		// Use default settings
		g.Model.SetTemperature(g.Temperature)
		g.Model.SetMaxOutputTokens(g.MaxTokens)
	}

	// Generate content
	resp, err := g.Model.GenerateContent(ctx, genPrompt...)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from the model")
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// ChatCompletion generates a response based on chat messages
func (g *GoogleGenAI) ChatCompletion(ctx context.Context, messages []Message) (Response, error) {
	if g.Client == nil || g.Model == nil {
		return Response{}, fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	var inputParts []genai.Part

	for _, msg := range messages {
		inputParts = append(inputParts, genai.Text(msg.Role+":"+msg.Content))
	}

	respGen := g.Model.GenerateContentStream(ctx, inputParts...)

	var response Response
	var resultContent string

	for {
		res, err := respGen.Next()
		if err != nil {
			break // End of stream
		}

		if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
			continue
		}

		for _, part := range res.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				resultContent += string(text)
			}
		}
	}

	response.Content = resultContent
	response.Raw = resultContent

	return response, nil
}

// ChatCompletionWithTools generates a response with tools support
func (g *GoogleGenAI) ChatCompletionWithTools(ctx context.Context, messages []Message, tools []Tool, options map[string]interface{}) (Response, error) {
	if g.Client == nil || g.Model == nil {
		return Response{}, fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	fmt.Printf("LLM Request - sapiens: %+v\n", messages)

	// Convert sapiens messages to genai chat messages
	var inputParts []genai.Part

	for _, msg := range messages {
		inputParts = append(inputParts, genai.Text(msg.Role+":"+msg.Content))
	}
	// Create tool definitions for the model
	var toolDefs []*genai.Tool
	for _, tool := range tools {
		var schema *genai.Schema

		if tool.InputSchema != nil {
			schema = &genai.Schema{
				Type:       genai.TypeObject,
				Properties: make(map[string]*genai.Schema),
				Required:   tool.InputSchema.Required,
			}

			for key, prop := range tool.InputSchema.Properties {

				var propType genai.Type
				switch prop.Type {
				case "string":
					propType = genai.TypeString
				case "number":
					propType = genai.TypeNumber
				case "boolean":
					propType = genai.TypeBoolean
				case "array":
					propType = genai.TypeArray
				case "object":
					propType = genai.TypeObject
				default:
					propType = genai.TypeString
				}

				schema.Properties[key] = &genai.Schema{
					Type:        propType,
					Description: prop.Description,
					Enum:        prop.Enum,
				}
			}
		}

		toolDefs = append(toolDefs, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  schema,
				},
			},
		})
	}

	// Set up tools and parameters
	g.Model.Tools = toolDefs
	g.Model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAuto,
		},
	}

	// Apply options
	if options != nil {
		if temp, ok := options["temperature"].(float32); ok {
			g.Model.SetTemperature(temp)
		}
		if maxTokens, ok := options["max_tokens"].(int32); ok {
			g.Model.SetMaxOutputTokens(maxTokens)
		}
	} else {
		// Use default settings
		g.Model.SetTemperature(g.Temperature)
		g.Model.SetMaxOutputTokens(g.MaxTokens)
	}

	// Generate content
	resp, err := g.Model.GenerateContent(ctx, inputParts...)

	if err != nil {
		return Response{}, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return Response{}, fmt.Errorf("no candidates returned from the model")
	}

	// Process response including possible tool calls
	response := Response{
		Raw: resp,
	}

	fmt.Printf("LLM Response - ai: %+v\n", resp.Candidates[0].Content.Parts)

	// Extract content and potential tool calls
	for _, part := range resp.Candidates[0].Content.Parts {
		switch p := part.(type) {
		case genai.Text:
			response.Content += string(p)
		case genai.FunctionCall:
			toolCall := ToolCall{
				ID:       p.Name, // Using name as ID as genai might not provide explicit IDs
				Name:     p.Name,
				InputMap: p.Args,
			}
			response.ToolCalls = append(response.ToolCalls, toolCall)
		}
	}

	return response, nil
}

// ChatCompletionWithToolsAndHandlers generates a response with tools support and executes tool handlers
func (g *GoogleGenAI) ChatCompletionWithToolsAndHandlers(ctx context.Context, messages []Message, tools []Tool,
	toolHandlers map[string]ToolHandler, options map[string]interface{}) (Response, error) {

	// Get initial response with tool calls
	response, err := g.ChatCompletionWithTools(ctx, messages, tools, options)
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
		finalResponse, err := g.ChatCompletion(ctx, newMessages)
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
func (g *GoogleGenAI) StructuredOutput(ctx context.Context, messages []Message, schema Schema) (Response, error) {
	if g.Client == nil || g.Model == nil {
		return Response{}, fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	// Add schema instruction to the last message or create a new one
	var schemaInstruction string
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal schema: %v", err)
	}

	schemaInstruction = "Please provide a response that conforms to the following JSON schema:\n```json\n" +
		string(schemaBytes) + "\n```\nYour response should be valid JSON that follows this schema."

	modifiedMessages := make([]Message, len(messages))
	copy(modifiedMessages, messages)

	// Append schema instruction to the last user message or add a new message
	if len(modifiedMessages) > 0 && modifiedMessages[len(modifiedMessages)-1].Role == "user" {
		modifiedMessages[len(modifiedMessages)-1].Content += "\n\n" + schemaInstruction
	} else {
		modifiedMessages = append(modifiedMessages, Message{
			Role:    "user",
			Content: schemaInstruction,
		})
	}

	// Get the completion
	responseWithSchema, err := g.ChatCompletion(ctx, modifiedMessages)
	if err != nil {
		return Response{}, err
	}

	// Extract JSON from the response
	jsonStr := responseWithSchema.Content

	// Try to parse the response as JSON
	var structuredData interface{}
	err = json.Unmarshal([]byte(jsonStr), &structuredData)
	if err != nil {
		return Response{Content: jsonStr}, fmt.Errorf("response is not valid JSON: %v", err)
	}

	return Response{
		Content:    jsonStr,
		Structured: structuredData,
		Raw:        responseWithSchema.Raw,
	}, nil
}
