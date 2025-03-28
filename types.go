package sapiens

import (
	"context"

	"github.com/patrickmn/go-cache"
)

type Vector []float64

// Prompt defines the system prompt and its version.
type Prompt struct {
	Prompt   string   `json:"prompt"`
	Version  string   `json:"version"`
	Enhanced []Prompt `json:"enhanced"`
}

// Tool represents an available tool for the agent.
type Tool struct {
	Name          string  `json:"name"`           // name of the tool
	Description   string  `json:"description"`    // describe the job of this tool and when do you want it to run.
	InputSchema   *Schema `json:"input_schema"`   // define parameters that are required for this tool
	OutputSchema  *Schema `json:"output_schema"`  // define the output of this tool
	RequiredTools []Tool  `json:"required_tools"` // list of tools that are required to run this tool
	Cost          float64 `json:"cost"`           // cost of running this tool
}

// Schema represents a JSON schema.
type Schema struct {
	Type        string            `json:"type"`        // string, number, integer, boolean, object, array
	Format      string            `json:"format"`      // markdown, json, xml
	Description string            `json:"description"` // description of the condition when this task is supposed to run
	Nullable    bool              `json:"nullable"`    // whether the value can be null
	Enum        []string          `json:"enum"`        // list of possible values
	Items       *Schema           `json:"items"`       // for array types
	Properties  map[string]Schema `json:"properties"`  // for object types
	Required    []string          `json:"required"`    // for object types
}

type Memory struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	Store  *cache.Cache           `json:"store"`
}

type EmbeddingType string

type Embedding struct {
	LLM     LLMInterface    // LLM used for the embedding
	Context context.Context // Context of the embedding
	Model   string          // Model used for the embedding
	Vector  []float64       // Embedding vector
	Text    string          // Original text
	Type    EmbeddingType   // Type of embedding
}

type SimilarityResult struct {
	Text      string      `json:"text"`
	Score     float64     `json:"score"`
	Embedding Embedding   `json:"embedding"`
	Key       interface{} `json:"key"` // Add the Key field
}

// LLM defines the basic information for a language model.
type LLM struct {
	Implementation LLMInterface
	ApiKey         string
	Model          string
	Provider       string
}

// Message represents a chat message in a conversation.
type Message struct {
	Role       string                 `json:"role"`    // system, user, assistant, or function
	Content    string                 `json:"content"` // message content
	Name       string                 `json:"name,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"` // ID of the tool call this message is responding to
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`   // Tool calls made by the assistant
	Options    map[string]interface{} `json:"options,omitempty"`      // additional model-specific options
}

// ToolCall represents a call to a tool by the LLM.
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Input    string                 `json:"input,omitempty"`     // String representation of input parameters
	InputMap map[string]interface{} `json:"input_map,omitempty"` // Structured input parameters
}

// Response represents a response from an LLM.
type Response struct {
	Content     string      `json:"content"`
	ToolCalls   []ToolCall  `json:"tool_calls,omitempty"`
	Structured  interface{} `json:"structured,omitempty"`   // Structured data based on schema
	ToolResults []Message   `json:"tool_results,omitempty"` // Results from tool execution
	Raw         interface{} `json:"raw,omitempty"`          // Raw response from the LLM
}

type AgentType string

// Agent represents an AI agent that can process queries and use tools
type Agent struct {
	Name                     string
	Type                     AgentType // Type of agent
	LLM                      *LLM
	SystemPrompts            []SystemPrompt
	StructuredResponseSchema Schema
	Tools                    []Tool
	toolImplementations      map[string]ToolImplementation
	Messages                 []Message
	conversationHistory      []Message
	MaxRetry                 int
	Context                  map[string]interface{}
	MetaData                 map[string]interface{}
	PromptManager            *PromptManager
	Memory                   *Memory
}

// SystemPrompt represents a system prompt with content and version
type SystemPrompt struct {
	Content string
	Version string
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"`
}

// Request represents a request to the LLM
type Request struct {
	Messages                 []Message      `json:"messages"`
	Tools                    []Tool         `json:"tools,omitempty"`
	StructuredResponseSchema *Schema        `json:"structured_response_schema,omitempty"`
	SystemPrompts            []SystemPrompt `json:"system_prompts,omitempty"`
}
