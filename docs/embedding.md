# Embedding

The `Embedding` type represents a vector embedding of text content.

## Structure

```go
type Embedding struct {
	Vector Vector // Vector representation
	Text   string // Original text
}
```

## Usage

```go
// Create a new embedding
embedding := sapiens.NewEmbedding(agent, "This is a sample text to embed")

// Use the embedding with memory
memory.Add("sample", data, embedding)

// Search for similar embeddings
results := memory.Search(embedding.Vector)
```

## Methods

### NewEmbedding

```go
func NewEmbedding(agent Agent, text string) Embedding
```

Creates a new Embedding for the provided text using the agent's embedding capabilities.

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Vector | Vector | Vector representation of the text |
| Text | string | Original text that was embedded |

## Related Types

- [Vector](vector.md)
- [Memory](memory.md)
- [Agent](agent.md)
