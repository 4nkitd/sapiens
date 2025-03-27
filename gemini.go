package sapiens

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai" // Import OpenAI
	"google.golang.org/api/option"
)

// GoogleGenAI implements the LLMInterface for Google's Generative AI
type GoogleGenAI struct {
	Client         *genai.Client
	SystemPrompt   SystemPrompt
	Model          *genai.GenerativeModel
	APIKey         string
	ModelName      string
	MaxTokens      int32
	Temperature    float32
	EmbeddingModel *genai.EmbeddingModel
}

func (g *GoogleGenAI) GenerateEmbedding(ctx context.Context, model string, text string, embeddingType EmbeddingType) (Embedding, error) {

	em := g.EmbeddingModel

	// Set the TaskType based on the embeddingType
	switch embeddingType {
	case "SEMANTIC_SIMILARITY":
		em.TaskType = genai.TaskTypeSemanticSimilarity
	case "CLASSIFICATION":
		em.TaskType = genai.TaskTypeClassification
	case "CLUSTERING":
		em.TaskType = genai.TaskTypeClustering
	case "RETRIEVAL_DOCUMENT":
		em.TaskType = genai.TaskTypeRetrievalDocument
	case "RETRIEVAL_QUERY":
		em.TaskType = genai.TaskTypeRetrievalQuery
	case "QUESTION_ANSWERING":
		em.TaskType = genai.TaskTypeQuestionAnswering
	case "FACT_VERIFICATION":
		em.TaskType = genai.TaskTypeFactVerification
	default:
		return Embedding{}, fmt.Errorf("unsupported embedding type: %s", embeddingType)
	}

	res, err := em.EmbedContent(ctx, genai.Text(text))

	if err != nil {
		return Embedding{}, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if res.Embedding == nil || len(res.Embedding.Values) == 0 {
		return Embedding{}, fmt.Errorf("embedding generation returned an empty vector")
	}

	// Convert []float32 to []float64
	float64Values := make([]float64, len(res.Embedding.Values))
	for i, v := range res.Embedding.Values {
		float64Values[i] = float64(v)
	}

	return Embedding{
		Vector: float64Values,
		Text:   text,
		Type:   embeddingType,
	}, nil
}

func (o *GoogleGenAI) GetModelName() string {
	return o.ModelName
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
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
		return nil // or panic, depending on your error handling strategy
	}

	embeddingModel := client.EmbeddingModel("gemini-embedding-exp-03-07")

	return &GoogleGenAI{
		Client:         client,
		APIKey:         apiKey,
		ModelName:      modelName,
		MaxTokens:      1024, // Default max tokens
		Temperature:    0.7,  // Default temperature
		EmbeddingModel: embeddingModel,
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

// SetSystemPrompt sets the system prompt for the model
func (g *GoogleGenAI) SetSystemPrompt(prompt SystemPrompt) {
	g.SystemPrompt = prompt
}

// CompleteWithOptions generates a completion with specific parameters
func (g *GoogleGenAI) CompleteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	if g.Client == nil || g.Model == nil {
		return "", fmt.Errorf("GoogleGenAI client not initialized, call Initialize() first")
	}

	// Create a prompt for the model
	inputParts := []genai.Part{
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

	if g.SystemPrompt.Content != "" {
		g.Model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(g.SystemPrompt.Content),
			},
			Role: "system",
		}
	}

	// Generate content
	resp, err := g.Model.GenerateContent(ctx, inputParts...)
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

	if g.SystemPrompt.Content != "" {
		g.Model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(g.SystemPrompt.Content),
			},
			Role: "system",
		}
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

	// Add special instruction to the system prompt for structured output if needed
	originalSystemPrompt := g.SystemPrompt.Content
	structuredToolNeeded := false

	for _, tool := range tools {
		if tool.Name == "structured_output" {
			structuredToolNeeded = true
			break
		}
	}

	// If we need structured output, adjust the system prompt
	if structuredToolNeeded {
		if originalSystemPrompt != "" {
			g.SystemPrompt.Content += "\nAfter calling any tools, please provide your final response in JSON format."
		} else {
			g.SystemPrompt.Content = "After calling any tools, please provide your final response in JSON format."
		}
	}

	// Apply options and set system instruction
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

	if g.SystemPrompt.Content != "" {
		g.Model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(g.SystemPrompt.Content),
			},
			Role: "system",
		}
	}

	// Generate content
	resp, err := g.Model.GenerateContent(ctx, inputParts...)

	// Restore original system prompt
	g.SystemPrompt.Content = originalSystemPrompt
	if originalSystemPrompt != "" {
		g.Model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(originalSystemPrompt),
			},
			Role: "system",
		}
	}

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

			if response.ToolCalls == nil {
				response.ToolCalls = make([]ToolCall, 0)
			}
			response.ToolCalls = append(response.ToolCalls, toolCall)

			// If this is a structured_output tool call, capture it as structured data
			if p.Name == "structured_output" {
				response.Structured = p.Args
			}
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

	// return direct in case of structured_output response
	if len(response.ToolCalls) == 1 && response.ToolCalls[0].Name == "structured_output" {
		return response, nil
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

	// Instead of using ResponseSchema which causes MIME type errors,
	// we'll add a direct instruction to return JSON
	// ConvertSchema, _ := (&schema).ConvertSchema(schema)

	// Create instruction for structured output - make this more direct and forceful
	schemaDescription := ""
	for name, prop := range schema.Properties {
		schemaDescription += fmt.Sprintf("- %s (%s): %s\n", name, prop.Type, prop.Description)
	}

	jsonInstruction := fmt.Sprintf(
		"You MUST respond with ONLY a valid JSON object containing these fields:\n%s\n"+
			"Your response MUST be valid JSON without any additional text, markdown formatting, or explanation.",
		schemaDescription,
	)

	originalSystemPrompt := g.SystemPrompt.Content
	// Set system prompt with JSON instruction
	if originalSystemPrompt != "" {
		g.SystemPrompt.Content = originalSystemPrompt + "\n\n" + jsonInstruction
	} else {
		g.SystemPrompt.Content = jsonInstruction
	}

	// Set the system instruction
	g.Model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(g.SystemPrompt.Content),
		},
		Role: "system",
	}

	// Ensure ResponseSchema is not set
	g.Model.ResponseSchema = nil

	// Process messages into genai format
	var inputParts []genai.Part
	for _, msg := range messages {
		inputParts = append(inputParts, genai.Text(msg.Role+":"+msg.Content))
	}

	// Generate content directly
	resp, err := g.Model.GenerateContent(ctx, inputParts...)

	// Restore original system prompt
	g.SystemPrompt.Content = originalSystemPrompt
	if originalSystemPrompt != "" {
		g.Model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(originalSystemPrompt),
			},
			Role: "system",
		}
	} else {
		g.Model.SystemInstruction = nil
	}

	if err != nil {
		return Response{}, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return Response{}, fmt.Errorf("no content in response")
	}

	// Extract content
	var content string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			content += string(text)
		}
	}

	// Try to parse the content as JSON
	var structured interface{}
	err = json.Unmarshal([]byte(content), &structured)
	if err != nil {
		// If it's not valid JSON, try to extract JSON from the response
		// Look for content between ``` markers
		jsonStart := strings.Index(content, "```json")
		if jsonStart != -1 {
			jsonStart += 7 // Skip ```json
			jsonEnd := strings.Index(content[jsonStart:], "```")
			if jsonEnd != -1 {
				// Extract JSON content
				jsonContent := strings.TrimSpace(content[jsonStart : jsonStart+jsonEnd])
				// Try to parse again
				err = json.Unmarshal([]byte(jsonContent), &structured)
				if err == nil {
					// We successfully parsed JSON from markdown code block
					return Response{
						Content:    content,
						Structured: structured,
						Raw:        resp,
					}, nil
				}
			}
		}

		// Return the content even if it's not structured properly
		return Response{Content: content, Raw: resp}, nil
	}

	return Response{
		Content:    content,
		Structured: structured,
		Raw:        resp,
	}, nil
}
