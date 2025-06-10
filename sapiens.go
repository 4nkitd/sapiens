package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {

	llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := NewAgent(context.Background(), llm.Client(), llm.GetDefaultModel(),
		"you are a weather reporter")

	message := NewMessages()

	// type Result struct {
	// 	Steps []struct {
	// 		Explanation string `json:"explanation"`
	// 		Output      string `json:"output"`
	// 	} `json:"steps"`
	// 	FinalAnswer string `json:"final_answer"`
	// }
	// var result Result

	// agent.SetResponseSchema("get_weather", "using this user can get wather details", true, result)
	//
	agent.AddTool("get_temp",
		"you can call this tool to get current temp",
		map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "The city and state, e.g. San Francisco, CA",
			},
			"unit": {
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
		},
		[]string{"location"},
		func(parameters map[string]string) string {

			for key, param := range parameters {
				fmt.Println("param-k:", key)
				fmt.Println("param-v:", param)
			}

			return `{"temperature":"27", "unit":"celsius"}`

		})

	resp, err := agent.Ask(message.MergeMessages(
		message.UserMessage("Can you tell me about the current weather conditions in Delhi? I need to know if I should wear a jacket today."),
	))

	if err != nil {
		log.Fatalf("CreateChatCompletion error: %v", err)
	}

	fmt.Println("AI Response:", resp.Choices[0].Message.Content)

}
