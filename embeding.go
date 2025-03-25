package sapiens

// NewEmbedding creates a new embedding for text
func NewEmbedding(agent Agent, text string) Embedding {
	return Embedding{
		Vector: []float32{},
		Text:   text,
	}
}
