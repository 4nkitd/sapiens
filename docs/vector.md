# Vector

The `Vector` type represents a numeric vector used for embeddings.

## Structure

```go
type Vector []float32
```

## Usage

```go
// Create a vector
vector := sapiens.Vector{0.1, 0.2, 0.3, 0.4}

// Create an embedding with a vector
embedding := sapiens.Embedding{
    Vector: vector,
    Text: "Sample text",
}

// Use in memory search
results := memory.Search(vector)
```

## Fields

A Vector is simply a slice of float32 values representing coordinates in a multi-dimensional space.

## Related Types

- [Embedding](embedding.md)
- [Memory](memory.md)
- [SimilarityResult](similarity-result.md)
