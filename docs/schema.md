# Schemas

Schemas in Sapiens define the structure and validation rules for tool parameters and structured responses. They use JSON Schema format and integrate with the `github.com/sashabaranov/go-openai/jsonschema` package for automatic generation from Go types.

## Overview

Schemas are used in two main areas:
1. **Tool Parameters** - Define what parameters tools accept
2. **Structured Responses** - Define the format of agent responses

## JSON Schema Definitions

### Basic Types

#### String
```go
"parameter_name": {
    Type:        jsonschema.String,
    Description: "A text parameter",
}
```

#### Number
```go
"amount": {
    Type:        jsonschema.Number,
    Description: "A numeric value",
}
```

#### Boolean
```go
"enabled": {
    Type:        jsonschema.Boolean,
    Description: "A true/false value",
}
```

#### Array
```go
"tags": {
    Type:        jsonschema.Array,
    Description: "A list of items",
    Items: &jsonschema.Definition{
        Type: jsonschema.String,
    },
}
```

#### Object
```go
"metadata": {
    Type:        jsonschema.Object,
    Description: "A nested object",
    Properties: map[string]jsonschema.Definition{
        "key": {Type: jsonschema.String},
        "value": {Type: jsonschema.String},
    },
}
```

### Enums (Limited Choices)

Restrict parameter values to a specific set of options:

```go
"priority": {
    Type:        jsonschema.String,
    Enum:        []string{"low", "medium", "high", "urgent"},
    Description: "Task priority level",
}

"temperature_unit": {
    Type:        jsonschema.String,
    Enum:        []string{"celsius", "fahrenheit", "kelvin"},
    Description: "Temperature measurement unit",
}
```

## Tool Parameter Schemas

Define what parameters your tools accept:

### Simple Tool Schema

```go
import "github.com/sashabaranov/go-openai/jsonschema"

agent.AddTool(
    "get_weather",
    "Get current weather for a location",
    map[string]jsonschema.Definition{
        "location": {
            Type:        jsonschema.String,
            Description: "City and country (e.g., 'London, UK')",
        },
        "unit": {
            Type:        jsonschema.String,
            Enum:        []string{"celsius", "fahrenheit"},
            Description: "Temperature unit",
        },
    },
    []string{"location"}, // Required parameters
    weatherToolFunc,
)
```

### Complex Tool Schema

```go
agent.AddTool(
    "create_task",
    "Create a new task in the project management system",
    map[string]jsonschema.Definition{
        "title": {
            Type:        jsonschema.String,
            Description: "Task title",
        },
        "description": {
            Type:        jsonschema.String,
            Description: "Detailed task description",
        },
        "priority": {
            Type:        jsonschema.String,
            Enum:        []string{"low", "medium", "high", "urgent"},
            Description: "Task priority",
        },
        "due_date": {
            Type:        jsonschema.String,
            Description: "Due date in YYYY-MM-DD format",
        },
        "assignee": {
            Type:        jsonschema.String,
            Description: "Email of the person to assign the task to",
        },
        "tags": {
            Type:        jsonschema.Array,
            Description: "List of tags for categorization",
            Items: &jsonschema.Definition{
                Type: jsonschema.String,
            },
        },
        "estimated_hours": {
            Type:        jsonschema.Number,
            Description: "Estimated hours to complete",
        },
    },
    []string{"title", "description"}, // Only title and description are required
    createTaskFunc,
)
```

## Structured Response Schemas

Define the format for structured agent responses using Go structs:

### Simple Structured Response

```go
// Define the response structure
type WeatherResponse struct {
    Temperature string `json:"temperature"`
    Condition   string `json:"condition"`
    Humidity    string `json:"humidity"`
    Location    string `json:"location"`
}

var weatherResponse WeatherResponse

// Set the schema on the agent
agent.SetResponseSchema(
    "weather_response",
    "Structured weather information",
    true, // strict validation
    weatherResponse,
)
```

### Complex Structured Response

```go
type AnalysisResult struct {
    Summary     string  `json:"summary"`
    Confidence  float64 `json:"confidence"`
    Steps       []Step  `json:"steps"`
    Metadata    Metadata `json:"metadata"`
}

type Step struct {
    StepNumber  int    `json:"step_number"`
    Description string `json:"description"`
    Result      string `json:"result"`
    Success     bool   `json:"success"`
}

type Metadata struct {
    ProcessingTime string            `json:"processing_time"`
    ToolsUsed      []string          `json:"tools_used"`
    AdditionalInfo map[string]string `json:"additional_info"`
}

var analysisResult AnalysisResult

agent.SetResponseSchema(
    "analysis_result",
    "Detailed analysis with steps and metadata",
    true,
    analysisResult,
)
```

### Working with Structured Responses

```go
// After getting a response from the agent
resp, err := agent.Ask(messages)
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Parse the structured response
var result AnalysisResult
err = agent.ParseResponse(resp, &result)
if err != nil {
    log.Fatalf("Parse error: %v", err)
}

// Access structured data
fmt.Printf("Summary: %s\n", result.Summary)
fmt.Printf("Confidence: %.2f\n", result.Confidence)
for _, step := range result.Steps {
    fmt.Printf("Step %d: %s -> %s\n", step.StepNumber, step.Description, step.Result)
}
```

## Schema Best Practices

### 1. Clear Descriptions

Always provide clear, helpful descriptions:

```go
// Good
"email": {
    Type:        jsonschema.String,
    Description: "Valid email address in format user@example.com",
}

// Better
"email": {
    Type:        jsonschema.String,
    Description: "Valid email address for notifications (e.g., john.doe@company.com)",
}
```

### 2. Use Enums for Limited Options

When parameters have limited valid values, use enums:

```go
"status": {
    Type:        jsonschema.String,
    Enum:        []string{"pending", "in_progress", "completed", "cancelled"},
    Description: "Current status of the task",
}
```

### 3. Specify Required Parameters

Always specify which parameters are required:

```go
agent.AddTool(
    "send_email",
    "Send an email message",
    parameters,
    []string{"to", "subject", "body"}, // Required
    emailFunc,
)
```

### 4. Use Appropriate Types

Choose the correct JSON schema type:

```go
// For counts, IDs, quantities
"quantity": {Type: jsonschema.Number}

// For text, names, descriptions
"name": {Type: jsonschema.String}

// For true/false values
"is_active": {Type: jsonschema.Boolean}

// For lists
"items": {Type: jsonschema.Array}
```

## Advanced Schema Examples

### File Operations Tool

```go
agent.AddTool(
    "file_operations",
    "Perform file system operations",
    map[string]jsonschema.Definition{
        "operation": {
            Type:        jsonschema.String,
            Enum:        []string{"read", "write", "delete", "list", "move"},
            Description: "Type of file operation to perform",
        },
        "path": {
            Type:        jsonschema.String,
            Description: "File or directory path",
        },
        "content": {
            Type:        jsonschema.String,
            Description: "Content to write (only for write operation)",
        },
        "new_path": {
            Type:        jsonschema.String,
            Description: "New path for move operation",
        },
        "recursive": {
            Type:        jsonschema.Boolean,
            Description: "Whether to perform operation recursively",
        },
    },
    []string{"operation", "path"},
    fileOperationsFunc,
)
```

### Database Query Tool

```go
agent.AddTool(
    "database_query",
    "Execute database queries",
    map[string]jsonschema.Definition{
        "query_type": {
            Type:        jsonschema.String,
            Enum:        []string{"select", "insert", "update", "delete"},
            Description: "Type of SQL query",
        },
        "table": {
            Type:        jsonschema.String,
            Description: "Database table name",
        },
        "conditions": {
            Type:        jsonschema.Object,
            Description: "Query conditions as key-value pairs",
        },
        "fields": {
            Type:        jsonschema.Array,
            Description: "Fields to select or update",
            Items: &jsonschema.Definition{
                Type: jsonschema.String,
            },
        },
        "limit": {
            Type:        jsonschema.Number,
            Description: "Maximum number of results",
        },
    },
    []string{"query_type", "table"},
    databaseQueryFunc,
)
```

### API Integration Tool

```go
agent.AddTool(
    "api_call",
    "Make HTTP API calls",
    map[string]jsonschema.Definition{
        "method": {
            Type:        jsonschema.String,
            Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
            Description: "HTTP method",
        },
        "url": {
            Type:        jsonschema.String,
            Description: "API endpoint URL",
        },
        "headers": {
            Type:        jsonschema.Object,
            Description: "HTTP headers as key-value pairs",
        },
        "body": {
            Type:        jsonschema.String,
            Description: "Request body (JSON string)",
        },
        "timeout": {
            Type:        jsonschema.Number,
            Description: "Request timeout in seconds",
        },
    },
    []string{"method", "url"},
    apiCallFunc,
)
```

## JSON Schema Generation

Sapiens automatically generates JSON schemas from Go structs for structured responses:

```go
// This struct
type UserProfile struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    Active   bool     `json:"active"`
    Tags     []string `json:"tags"`
    Metadata map[string]interface{} `json:"metadata"`
}

// Automatically becomes this schema:
// {
//   "type": "object",
//   "properties": {
//     "id": {"type": "integer"},
//     "name": {"type": "string"},
//     "email": {"type": "string"},
//     "active": {"type": "boolean"},
//     "tags": {
//       "type": "array",
//       "items": {"type": "string"}
//     },
//     "metadata": {"type": "object"}
//   },
//   "required": ["id", "name", "email", "active", "tags", "metadata"]
// }
```

## Error Handling

Handle schema-related errors gracefully:

```go
// When parsing structured responses
var result MyStruct
err := agent.ParseResponse(resp, &result)
if err != nil {
    log.Printf("Schema parsing error: %v", err)
    // Fall back to plain text response
    fmt.Println("Plain response:", resp.Choices[0].Message.Content)
    return
}

// When setting response schemas
schema := agent.SetResponseSchema("my_schema", "Description", true, result)
if schema == nil {
    log.Printf("Failed to set response schema")
    // Continue without structured output
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

type TaskResult struct {
    TaskID      string   `json:"task_id"`
    Status      string   `json:"status"`
    Description string   `json:"description"`
    Tags        []string `json:"tags"`
    CreatedAt   string   `json:"created_at"`
}

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a task management assistant",
    )
    
    // Add tool with complex schema
    agent.AddTool(
        "create_task",
        "Create a new task",
        map[string]jsonschema.Definition{
            "title": {
                Type:        jsonschema.String,
                Description: "Task title",
            },
            "priority": {
                Type:        jsonschema.String,
                Enum:        []string{"low", "medium", "high"},
                Description: "Task priority",
            },
            "tags": {
                Type:        jsonschema.Array,
                Description: "Task tags",
                Items: &jsonschema.Definition{Type: jsonschema.String},
            },
        },
        []string{"title"},
        func(params map[string]string) string {
            return `{
                "task_id": "task_123",
                "status": "created",
                "description": "Task created successfully",
                "tags": ["work", "urgent"],
                "created_at": "2024-01-15T10:30:00Z"
            }`
        },
    )
    
    // Set structured response schema
    var taskResult TaskResult
    agent.SetResponseSchema("task_result", "Task operation result", true, taskResult)
    
    // Use the agent
    message := NewMessages()
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Create a high priority task called 'Review documentation'"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Parse structured response
    err = agent.ParseResponse(resp, &taskResult)
    if err != nil {
        log.Printf("Parse error: %v", err)
        fmt.Println("Plain response:", resp.Choices[0].Message.Content)
    } else {
        fmt.Printf("Task created: %s (ID: %s, Status: %s)\n", 
            taskResult.Description, taskResult.TaskID, taskResult.Status)
    }
}
```

## Related Documentation

- [Tools](tool.md) - Using schemas with tool parameters
- [Agent](agent.md) - Setting up structured responses
- [Examples](examples.md) - More schema examples and patterns