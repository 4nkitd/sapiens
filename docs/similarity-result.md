# SimilarityResult

The `SimilarityResult` type represents the result of a similarity search in memory.

## Structure

```go
type SimilarityResult struct {
	Text      string    // Original text
	Score     float32   // Similarity score
	Embedding Embedding // The embedding that was matched
}
```

## Usage

```go
// Search memory for similar embeddings
results := memory.Search(queryEmbedding)

// Process search results
for _, result := range results {
    fmt.Printf("Match: %s (Score: %.2f)\n", result.Text, result.Score)
    
    // Use the matched embedding
    matchedEmbedding := result.Embedding
}
```

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Text | string | Original text associated with the embedding |
| Score | float32 | Similarity score (higher is more similar) |
| Embedding | Embedding | The embedding that was matched in the search |

## Related Types

- [Embedding](embedding.md)
- [Memory](memory.md)
- [Vector](vector.md)
