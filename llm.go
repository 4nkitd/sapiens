package sapiens

import (
	"context"
	// Import OpenAI
)

// ToolHandler is a function that handles tool calls from the LLM
type ToolHandler func(toolName string, arguments map[string]interface{}) (string, error)

// LLMInterface defines the contract for language model implementations.
type LLMInterface interface {
	// Initialize sets up the language model with provided options
	Initialize() error

	SetSystemPrompt(prompt SystemPrompt)

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
