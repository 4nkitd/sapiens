package sapiens

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestValidateEmbeddingType(t *testing.T) {

	testCases := []struct {
		embeddingType EmbeddingType
		expected      bool
	}{
		{"SEMANTIC_SIMILARITY", true},
		{"CLASSIFICATION", true},
		{"CLUSTERING", true},
		{"RETRIEVAL_DOCUMENT", true},
		{"RETRIEVAL_QUERY", true},
		{"QUESTION_ANSWERING", true},
		{"FACT_VERIFICATION", true},
		{"CODE_RETRIEVAL_QUERY", true},
		{"INVALID_TYPE", false},
		{"", false},
	}

	for _, tc := range testCases {
		actual := ValidateEmbeddingType(tc.embeddingType)
		if actual != tc.expected {
			t.Errorf("For embedding type %s, expected %v, but got %v", tc.embeddingType, tc.expected, actual)
		}
	}
}

func TestNewEmbeddingType(t *testing.T) {
	validType := "SEMANTIC_SIMILARITY"
	invalidType := "INVALID_TYPE"

	_, err := NewEmbeddingType(EmbeddingType(validType))
	if err != nil {
		t.Errorf("Expected no error for valid type, but got %v", err)
	}

	_, err = NewEmbeddingType(EmbeddingType(invalidType))
	if err == nil {
		t.Errorf("Expected error for invalid type, but got nil")
	}
}

func TestEmbedding_GenerateEmbedding(t *testing.T) {
	// Skip the test if the Gemini API key is not set
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping Gemini test because GEMINI_API_KEY is not set")
	}

	testCases := []struct {
		name          string
		text          string
		embeddingType EmbeddingType
		expectedErr   bool
	}{
		{
			name:          "Successful embedding generation",
			text:          "test text",
			embeddingType: "SEMANTIC_SIMILARITY",
			expectedErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
			}

			llm := NewGoogleGenAI(apiKey, "gemini-2.0-flash") // Google GenAI
			// llm := NewOpenAI(apiKey, "gpt-3.5-turbo") // OpenAI

			ctx := context.Background()
			embeddinInstance := NewEmbedding(ctx, llm)

			result, err := embeddinInstance.GenerateEmbedding(tc.text, tc.embeddingType)

			if (err != nil) != tc.expectedErr {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err != nil)
			}

			if err == nil && len(result.Vector) == 0 {
				t.Errorf("Expected non-empty vector, but got empty vector")
			}
			if !reflect.DeepEqual(result.Text, tc.text) {
				t.Errorf("Expected text: %v, but got: %v", tc.text, result.Text)
			}
			if !reflect.DeepEqual(result.Type, tc.embeddingType) {
				t.Errorf("Expected type: %v, but got: %v", tc.embeddingType, result.Type)
			}
		})
	}
}
