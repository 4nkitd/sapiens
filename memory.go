package sapiens

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/patrickmn/go-cache"
)

// NewMemory creates a new memory instance
func NewMemory(memoryType string, config map[string]interface{}) Memory {
	return Memory{
		Type:   memoryType,
		Config: config,
		Store:  cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Add stores an item in memory with its embedding
func (m *Memory) Add(key string, value interface{}, embedding Embedding) {
	embeddingJSON, _ := json.Marshal(embedding)
	m.Store.Set(key, embeddingJSON, cache.DefaultExpiration)
}

// Get retrieves an item from memory by key
func (m *Memory) Get(key string) interface{} {
	if x, found := m.Store.Get(key); found {
		var embedding Embedding
		json.Unmarshal(x.([]byte), &embedding)
		return embedding
	}
	return nil
}

// Remove deletes an item from memory by key
func (m *Memory) Remove(key string) {
	m.Store.Delete(key)
}

// Reset clears all items from memory
func (m *Memory) Reset() {
	m.Store.Flush()
}

// cosineSimilarity calculates similarity between two vectors
func (m *Memory) cosineSimilarity(vec1, vec2 Vector) float64 {
	sum := float64(0.0)
	for i := 0; i < len(vec1); i++ {
		sum += float64(vec1[i] * vec2[i])
	}
	return sum
}

// Search finds similar embeddings ranked by similarity score
func (m *Memory) Search(queryEmbedding Vector) []SimilarityResult {
	var results []SimilarityResult

	for key, item := range m.Store.Items() { // Iterate with key
		var storedEmbedding Embedding
		json.Unmarshal(item.Object.([]byte), &storedEmbedding)

		similarity := m.cosineSimilarity(queryEmbedding, storedEmbedding.Vector)

		// Add to results
		results = append(results, SimilarityResult{
			Text:      storedEmbedding.Text,
			Score:     float64(similarity),
			Embedding: storedEmbedding,
			Key:       key, // Store the key
		})
	}

	// Sort results by similarity score in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
