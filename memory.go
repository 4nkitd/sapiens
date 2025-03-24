package sapiens

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/patrickmn/go-cache"
)

func NewMemory(memoryType string, config map[string]interface{}) Memory {
	return Memory{
		Type:   memoryType,
		Config: config,
		Store:  cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (m *Memory) Add(key string, value interface{}, embedding Embedding) {

	embedding1JSON, _ := json.Marshal(embedding)

	m.Store.Set(key, embedding1JSON, cache.DefaultExpiration)

}

func (m *Memory) Get(key string) interface{} {
	if x, found := m.Store.Get(key); found {
		var embedding Embedding
		json.Unmarshal(x.([]byte), &embedding)
		return embedding
	}
	return nil
}

func (m *Memory) Remove(key string) {
	m.Store.Delete(key)
}

func (m *Memory) Reset() {
	m.Store.Flush()
}

// Function to calculate cosine similarity.
func (m *Memory) cosineSimilarity(vec1, vec2 Vector) float32 {
	// Implement cosine similarity calculation here.
	// Example only, do not use in production.
	sum := float32(0.0)
	for i := 0; i < len(vec1); i++ {
		sum += vec1[i] * vec2[i]
	}
	return sum
}

// Function to find similar embeddings ranked by similarity score.
func (m *Memory) Search(queryEmbedding Vector) []SimilarityResult {
	var results []SimilarityResult

	for _, item := range m.Store.Items() {
		var storedEmbedding Embedding
		json.Unmarshal(item.Object.([]byte), &storedEmbedding)

		similarity := m.cosineSimilarity(queryEmbedding, storedEmbedding.Vector)

		// Add to results
		results = append(results, SimilarityResult{
			Text:      storedEmbedding.Text,
			Score:     similarity,
			Embedding: storedEmbedding,
		})
	}

	// Sort results by similarity score in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
