# Tool

The `Tool` type represents a capability or function that an agent can use.

## Structure

```go
type Tool struct {
	Name          string   // Name of the tool
	Description   string   // Description of the tool's purpose
	InputSchema   *Schema  // Schema for the tool's input parameters
	OutputSchema  *Schema  // Schema for the tool's output format
	RequiredTools []Tool   // Other tools required by this tool
	Cost          float64  // Cost associated with using this tool
}
```

## Usage

```go
// Create a new tool
tool := sapiens.NewTool(
    "weather_lookup",
    "Look up the current weather for a location",
    []sapiens.Tool{}
)

// Set cost
tool.SetCost(0.01)

// Define input schema
tool.AddInputSchema(&sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "location": {Type: "string"},
    },
    Required: []string{"location"},
})

// Define output schema
tool.AddOutputSchema(&sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "temperature": {Type: "number"},
        "condition": {Type: "string"},
    },
})
```

## Methods

### NewTool

```go
func NewTool(name string, description string, requiredTools []Tool) Tool
```

Creates a new Tool with the specified name, description, and required tools.

### SetCost

```go
func (t *Tool) SetCost(cost float64)
```

Sets the cost associated with using this tool.

### AddInputSchema

```go
func (t *Tool) AddInputSchema(inputFormat *Schema)
```

Adds an input schema to the tool.

### AddOutputSchema

```go
func (t *Tool) AddOutputSchema(outputFormat *Schema)
```

Adds an output schema to the tool.

### AddRequiredTool

```go
func (t *Tool) AddRequiredTool(tool Tool)
```

Adds a required tool dependency.

### SetDescription

```go
func (t *Tool) SetDescription(description string)
```

Sets the description of the tool.

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Name | string | Identifier for the tool |
| Description | string | Description of what the tool does |
| InputSchema | *Schema | Defines the required input format |
| OutputSchema | *Schema | Defines the expected output format |
| RequiredTools | []Tool | Other tools this tool depends on |
| Cost | float64 | Cost associated with using this tool |

## Related Types

- [Agent](agent.md)
- [Schema](schema.md)
