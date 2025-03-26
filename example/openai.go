package main

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	client := openai.NewClient(
		option.WithBaseURL("https://models.inference.ai.azure.com"),   // defaults to https://api.openai.com
		option.WithAPIKey("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"), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage("what is the capital of the United States?"),
			},
			Model: openai.ChatModelGPT4o,
		},
	)

	if err != nil {
		panic(err.Error())
	}

	println(chatCompletion.Choices[0].Message.Content)
}
