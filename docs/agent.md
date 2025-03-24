# Agent

The `Agent` type represents an AI agent with its configuration and capabilities.

## Structure

```go
type Agent struct {
	Name                     string                 // Name of the agent
	SystemPrompt             Prompt                 // System prompt that defines the agent's behavior
	Tools                    []Tool                 // Available tools the agent can use
	StructuredResponseSchema Schema                 // Schema for structured responses
	Memory                   []Memory               // Memory instances used by the agent
	MaxRetry                 int                    // Maximum number of retry attempts
	Context                  map[string]interface{} // Context information
	MetaData                 map[string]interface{} // Metadata about the agent
}
```

## Usage

```go
// Create a new agent
agent := sapiens.Agent{
    Name: "My Assistant",
    SystemPrompt: systemPrompt,
    Tools: []sapiens.Tool{},
    Memory: []sapiens.Memory{},
    MaxRetry: 3,
    Context: map[string]interface{}{
        "environment": "production",
    },
    MetaData: map[string]interface{}{
        "version": "1.0.0",
    },
}
```

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Name | string | Identifier for the agent |
| SystemPrompt | Prompt | Defines the agent's behavior and instructions |
| Tools | []Tool | Collection of tools the agent can use |
| StructuredResponseSchema | Schema | Defines the format for structured outputs |
| Memory | []Memory | Memory instances for storing information |
| MaxRetry | int | Maximum number of retry attempts for failed operations |
| Context | map[string]interface{} | Runtime context information |
| MetaData | map[string]interface{} | Additional agent metadata |

## Related Types

- [Prompt](prompt.md)
- [Tool](tool.md)
- [Schema](schema.md)
- [Memory](memory.md)
