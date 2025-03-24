# Prompt

The `Prompt` type represents a system prompt with versioning and enhancement capabilities.

## Structure

```go
type Prompt struct {
	Prompt   string   // The actual prompt text
	Version  string   // Version of the prompt
	Enhanced []Prompt // Collection of enhanced versions of this prompt
}
```

## Usage

```go
// Create a new system prompt
prompt := sapiens.NewSystemPrompt(
    "You are an AI assistant named {{.name}}.",
    "v1.0"
)

// Parse the prompt with parameters
params := map[string]string{"name": "Sapiens"}
parsedPrompt, err := prompt.Parse(prompt.Prompt, params)
```

## Methods

### NewSystemPrompt

```go
func NewSystemPrompt(prompt string, version string) Prompt
```

Creates a new Prompt with the specified text and version.

### SetEnhanced

```go
func (s *Prompt) SetEnhanced(enhanced Prompt)
```

Adds an enhanced version of the prompt.

### GetLatestEnhanced

```go
func (s *Prompt) GetLatestEnhanced() Prompt
```

Returns the most recent enhanced version of the prompt.

### Parse

```go
func (s *Prompt) Parse(raw string, parameters map[string]string) (output Prompt, err error)
```

Parses a prompt template with provided parameters to create a fully rendered prompt.

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Prompt | string | The actual prompt text |
| Version | string | Version identifier of the prompt |
| Enhanced | []Prompt | Collection of enhanced versions of this prompt |

## Related Types

- [Agent](agent.md)
