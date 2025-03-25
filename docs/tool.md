# Tool

The `Tool` type represents a capability or function that an agent can call.

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
    []sapiens.Tool{},
)

// Set cost
tool.SetCost(0.01)

// Define input schema
tool.AddInputSchema(&sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "location": {Type: "string"},
        "unit": {
            Type: "string",
            Enum: []string{"celsius", "fahrenheit"},
        },
    },
    Required: []string{"location"},
})

// Define output schema
tool.AddOutputSchema(&sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "temperature": {Type: "number"},
        "condition": {Type: "string"},
        "humidity": {Type: "number"},
        "wind_speed": {Type: "number"},
    },
})
```

## Tool Implementation

Tools are implemented by registering handler functions with the agent:

```go
// Add the tool to an agent
agent.AddTools(tool)

// Register the implementation
agent.RegisterToolImplementation("weather_lookup", func(params map[string]interface{}) (interface{}, error) {
    // Extract parameters
    location, ok := params["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location must be a string")
    }
    
    unit := "celsius"
    if unitParam, ok := params["unit"].(string); ok {
        unit = unitParam
    }
    
    // Implement tool functionality
    // (e.g., call a weather API)
    
    // Return results
    return map[string]interface{}{
        "temperature": 22.5,
        "condition": "Partly Cloudy",
        "humidity": 65,
        "wind_speed": 15,
        "unit": unit,
    }, nil
})
```

## Tool Calls

When the language model determines a tool needs to be called, it generates a `ToolCall`:

```go
type ToolCall struct {
	ID       string                 // Unique identifier for this tool call
	Name     string                 // Name of the tool to call
	Input    string                 // String representation of input parameters
	InputMap map[string]interface{} // Structured input parameters
}
```

The agent automatically handles tool calls by:
1. Receiving the tool call in the LLM response
2. Looking up the registered implementation for the tool
3. Executing the tool with the provided parameters
4. Adding the tool result to the conversation
5. Sending the conversation with the tool result back to the LLM

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

## Example Tool Types

### Calculator

```go
// Calculator tool
calculatorTool := sapiens.Tool{
    Name:        "calculator",
    Description: "Perform mathematical calculations",
    InputSchema: &sapiens.Schema{
        Type: "object",
        Properties: map[string]sapiens.Schema{
            "expression": {
                Type:        "string",
                Description: "Mathematical expression to evaluate",
            },
        },
        Required: []string{"expression"},
    },
}

// Implementation
agent.RegisterToolImplementation("calculator", func(params map[string]interface{}) (interface{}, error) {
    expression := params["expression"].(string)
    // Use a math expression evaluator
    result := evaluateMathExpression(expression)
    return map[string]interface{}{
        "result": result,
    }, nil
})
```

### Search

```go
// Search tool
searchTool := sapiens.Tool{
    Name:        "search",
    Description: "Search the web for information",
    InputSchema: &sapiens.Schema{
        Type: "object",
        Properties: map[string]sapiens.Schema{
            "query": {
                Type:        "string",
                Description: "Search query",
            },
            "num_results": {
                Type:        "integer",
                Description: "Number of results to return",
            },
        },
        Required: []string{"query"},
    },
}

// Implementation
agent.RegisterToolImplementation("search", func(params map[string]interface{}) (interface{}, error) {
    query := params["query"].(string)
    numResults := 3
    if n, ok := params["num_results"].(float64); ok {
        numResults = int(n)
    }
    
    results := performWebSearch(query, numResults)
    return map[string]interface{}{
        "results": results,
    }, nil
})
```

## Related Types

- [Agent](agent.md)
- [Schema](schema.md)
- [ToolCall](llm.md#tool-integration)
