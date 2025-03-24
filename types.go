package sapiens

import "github.com/patrickmn/go-cache"

type Vector []float32

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

type Embedding struct {
	Vector Vector `json:"vector"`
	Text   string `json:"text"`
}

type SimilarityResult struct {
	Text      string
	Score     float32
	Embedding Embedding
}

type Agent struct {
	Name                     string                 `json:"name"`
	SystemPrompt             Prompt                 `json:"system_prompt"`
	Tools                    []Tool                 `json:"tools"`
	StructuredResponseSchema Schema                 `json:"structured_response_schema"`
	Memory                   []Memory               `json:"memory"`
	MaxRetry                 int                    `json:"max_retry"`
	Context                  map[string]interface{} `json:"context"`
	MetaData                 map[string]interface{} `json:"meta_data"`
}
