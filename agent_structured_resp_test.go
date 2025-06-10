package sapiens

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestAgentStructuredResponse(t *testing.T) {

	llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := NewAgent(context.Background(), llm.Client(), llm.GetDefaultModel(),
		"you are a weather reporter")

	message := NewMessages()

	type Result struct {
		Steps []struct {
			Explanation string `json:"explanation"`
			Output      string `json:"output"`
		} `json:"steps"`
		FinalAnswer string `json:"final_answer"`
	}
	var result Result

	agent.SetResponseSchema("get_weather", "using this user can get wather details", true, result)

	resp, err := agent.Ask(message.MergeMessages(
		message.UserMessage("what can you do"),
	))

	if err != nil {
		log.Fatalf("CreateChatCompletion error: %v", err)
	}

	agent.ParseResponse(resp, &result)

	log.Printf("Unmarshalled result: %+v", result)

}
