# Memory

The `Memory` type provides storage and retrieval capabilities for agents.

## Structure

```go
type Memory struct {
	Type   string                 // Type of memory
	Config map[string]interface{} // Configuration parameters
	Store  *cache.Cache           // Actual storage implementation
}
```

## Usage

```go
// Create a new memory
memory := sapiens.NewMemory("semantic", map[string]interface{}{
    "dimension": 1536,
})

// Add an embedding to memory
embedding := sapiens.NewEmbedding(agent, "This is important information")
memory.Add("info1", data, embedding)

// Retrieve an item
result := memory.Get("info1")

// Search for similar items
results := memory.Search(queryEmbedding)

// Remove an item
memory.Remove("info1")

// Clear all memory
memory.Reset()
```

## Methods

### NewMemory

```go
func NewMemory(memoryType string, config map[string]interface{}) Memory
```

Creates a new Memory with the specified type and configuration.

### Add

```go
func (m *Memory) Add(key string, value interface{}, embedding Embedding)
```

Adds an item to memory with the specified key, value, and embedding.

### Get

```go
func (m *Memory) Get(key string) interface{}
```

Retrieves an item from memory by key.

### Remove

```go
func (m *Memory) Remove(key string)
```

Removes an item from memory by key.

### Reset

```go
func (m *Memory) Reset()
```

Clears all items from memory.

### Search

```go
func (m *Memory) Search(queryEmbedding Vector) []SimilarityResult
```

Searches for items similar to the provided query embedding.

## Fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| Type | string | Type of memory (e.g., "semantic", "episodic") |
| Config | map[string]interface{} | Configuration parameters for the memory |
| Store | *cache.Cache | Underlying storage implementation |

## Related Types

- [Agent](agent.md)
- [Embedding](embedding.md)
- [Vector](vector.md)
- [SimilarityResult](similarity-result.md)
