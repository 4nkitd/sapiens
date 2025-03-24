# Schema

The `Schema` type represents a JSON schema for defining the structure of data.

## Structure

```go
type Schema struct {
	Type        string            // Data type (string, number, integer, boolean, object, array)
	Format      string            // Format specifier (markdown, json, xml, etc.)
	Description string            // Description of the schema
	Nullable    bool              // Whether the value can be null
	Enum        []string          // List of possible values
	Items       *Schema           // For array types, defines the schema of items
	Properties  map[string]Schema // For object types, defines properties
	Required    []string          // For object types, list of required properties
}
```

## Usage

```go
// Create a schema for a user object
userSchema := sapiens.Schema{
    Type: "object",
    Description: "User information",
    Properties: map[string]sapiens.Schema{
        "name": {
            Type: "string",
            Description: "User's full name",
        },
        "age": {
            Type: "integer",
            Description: "User's age in years",
            Nullable: true,
        },
        "email": {
            Type: "string",
            Format: "email",
            Description: "User's email address",
        },
        "preferences": {
            Type: "array",
            Items: &sapiens.Schema{
                Type: "string",
                Enum: []string{"light", "dark", "system"},
            },
        },
    },
    Required: []string{"name", "email"},
}
```

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Type | string | Data type (string, number, integer, boolean, object, array) |
| Format | string | Format specifier (markdown, json, xml, etc.) |
| Description | string | Description of the schema or field |
| Nullable | bool | Whether the value can be null |
| Enum | []string | List of possible values |
| Items | *Schema | For array types, defines the schema of contained items |
| Properties | map[string]Schema | For object types, defines property schemas |
| Required | []string | For object types, list of required properties |

## Related Types

- [Tool](tool.md)
- [Agent](agent.md)
