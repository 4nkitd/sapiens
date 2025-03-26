package sapiens

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/invopop/jsonschema"
	openai "github.com/openai/openai-go"
)

// Property defines a property in a JSON schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// HistoricalComputer struct that will be converted to a Structured Outputs response schema
type HistoricalComputer struct {
	Origin       Origin   `json:"origin" jsonschema:"description=The origin of the computer"`
	Name         string   `json:"full_name" jsonschema:"description=The name of the device model"`
	Legacy       string   `json:"legacy" jsonschema:"enum=positive,enum=neutral,enum=negative,description=Its influence on the field of computing"`
	NotableFacts []string `json:"notable_facts" jsonschema:"description=A few key facts about the computer"`
}

// Origin struct
type Origin struct {
	YearBuilt    int64  `json:"year_of_construction" jsonschema:"description=The year it was made"`
	Organization string `json:"organization" jsonschema:"description=The organization that was in charge of its development"`
}

// GenerateSchema generates a JSON schema from a Go type
func GenerateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func TestOpenAI(t *testing.T) {

	os.Setenv("OPENAI_API_KEY", "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")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY environment variable not set")
	}

	modelName := "gpt-4o" // Replace with an appropriate model name

	// Create a new OpenAI implementation
	llmImpl := NewOpenAI(apiKey, modelName)
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize OpenAI: %v", err)
	}

	// Test basic completion
	t.Run("Complete", func(t *testing.T) {
		ctx := context.Background()
		response, err := llmImpl.Complete(ctx, "What is the capital of France?")
		if err != nil {
			t.Fatalf("Complete failed: %v", err)
		}
		t.Logf("Response: %s", response)
		if response == "" {
			t.Error("Response is empty")
		}
	})

	// Test chat completion
	t.Run("ChatCompletion", func(t *testing.T) {
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "What is the capital of Germany?"}}
		response, err := llmImpl.ChatCompletion(ctx, messages)
		if err != nil {
			t.Fatalf("ChatCompletion failed: %v", err)
		}
		t.Logf("Response: %s", response.Content)
		if response.Content == "" {
			t.Error("Response is empty")
		}
	})

	// Test chat completion with tools
	t.Run("ChatCompletionWithTools", func(t *testing.T) {
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "What is the weather in Berlin?"}}
		tools := []Tool{{
			Name:        "get_weather",
			Description: "Get weather at the given location",
			InputSchema: &Schema{
				Type:     "object",
				Required: []string{"location"},
				Properties: map[string]Schema{
					"location": {Type: "string", Description: "The city and state, e.g. San Francisco, CA"},
				},
			},
		}}
		response, err := llmImpl.ChatCompletionWithTools(ctx, messages, tools, nil)
		if err != nil {
			t.Fatalf("ChatCompletionWithTools failed: %v", err)
		}
		t.Logf("Response: %s", response.Content)
		// Add assertions to check for tool calls if expected
	})

	// Test chat completion with tools and handlers
	t.Run("ChatCompletionWithToolsAndHandlers", func(t *testing.T) {
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "What is the weather in London?"}}
		tools := []Tool{{
			Name:        "get_weather",
			Description: "Get weather at the given location",
			InputSchema: &Schema{
				Type:     "object",
				Required: []string{"location"},
				Properties: map[string]Schema{
					"location": {Type: "string", Description: "The city and state, e.g. San Francisco, CA"},
				},
			},
		}}
		toolHandlers := map[string]ToolHandler{
			"get_weather": func(toolName string, arguments map[string]interface{}) (string, error) {
				location, ok := arguments["location"].(string)
				if !ok {
					return "", nil
				}
				return location + " Sunny, 25C", nil
			},
		}
		response, err := llmImpl.ChatCompletionWithToolsAndHandlers(ctx, messages, tools, toolHandlers, nil)
		if err != nil {
			t.Fatalf("ChatCompletionWithToolsAndHandlers failed: %v", err)
		}
		t.Logf("Response: %s", response.Content)
		// Add assertions to check for handled tool calls and results
	})

	// Test structured output
	t.Run("StructuredOutput", func(t *testing.T) {
		ctx := context.Background()

		// Generate the JSON schema
		HistoricalComputerResponseSchema := GenerateSchema[HistoricalComputer]()

		schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        "historical_computer",
			Description: openai.String("Notable information about a computer"),
			Schema:      HistoricalComputerResponseSchema,
			Strict:      openai.Bool(true),
		}

		params := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage("What computer ran the first neural network?"),
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
			},
			Model: "gpt-4o",
		}

		resp, err := llmImpl.Client.Chat.Completions.New(ctx, params)
		if err != nil {
			t.Fatalf("StructuredOutput failed: %v", err)
		}

		if len(resp.Choices) == 0 {
			t.Fatalf("no choices returned from the model")
		}

		// The model responds with a JSON string, so parse it into a struct
		var historicalComputer HistoricalComputer
		err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &historicalComputer)
		if err != nil {
			t.Fatalf("failed to unmarshal structured data: %v", err)
		}

		t.Logf("Response: %v", historicalComputer)

		// Add assertions to check for structured output
	})

	// Test generate
	t.Run("Generate", func(t *testing.T) {
		ctx := context.Background()
		request := &Request{Messages: []Message{{Role: "user", Content: "Tell me a joke."}}}
		response, err := llmImpl.Generate(ctx, request)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		t.Logf("Response: %s", response.Content)
		if response.Content == "" {
			t.Error("Response is empty")
		}
	})
}
