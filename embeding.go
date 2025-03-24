package sapiens

func NewEmbedding(agent Agent, text string) Embedding {

	return Embedding{
		Vector: []float32{},
		Text:   text,
	}

}
