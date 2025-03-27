package sapiens

import (
	"context"
	"fmt"
)

func NewEmbedding(ctx context.Context, llm LLMInterface) Embedding {
	modelName := llm.GetModelName()
	return Embedding{
		Context: ctx,
		Model:   modelName,
		LLM:     llm,
	}
}

func NewEmbeddingType(embedding_type EmbeddingType) (EmbeddingType, error) {
	if ValidateEmbeddingType(EmbeddingType(embedding_type)) {
		return EmbeddingType(embedding_type), nil
	}
	return EmbeddingType(""), fmt.Errorf("invalid embedding type")
}

func ValidateEmbeddingType(embeddingType EmbeddingType) bool {

	switch embeddingType {
	case "SEMANTIC_SIMILARITY":
		return true
	case "CLASSIFICATION":
		return true
	case "CLUSTERING":
		return true
	case "RETRIEVAL_DOCUMENT":
		return true
	case "RETRIEVAL_QUERY":
		return true
	case "QUESTION_ANSWERING":
		return true
	case "FACT_VERIFICATION":
		return true
	case "CODE_RETRIEVAL_QUERY":
		return true
	default:
		return false
	}

}

func (e *Embedding) GenerateEmbedding(text string, embeddingType EmbeddingType) (Embedding, error) {

	embedding, errEmbedding := e.LLM.GenerateEmbedding(e.Context, e.Model, text, embeddingType)
	if errEmbedding != nil {
		return Embedding{}, errEmbedding
	}

	return embedding, nil

}
