package sapiens

import openai "github.com/sashabaranov/go-openai"

type Messages struct {
}

func NewMessages() *Messages {

	return &Messages{}

}

func (a *Messages) UserMessage(msg string) openai.ChatCompletionMessage {

	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	}

}

func (a *Messages) ToolMessage(id, name, msg string) openai.ChatCompletionMessage {

	return openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    msg, // The string result from our function
		ToolCallID: id,  // The ID from the model's request
		Name:       name,
	}
}

func (a *Messages) AgentMessage(msg string) openai.ChatCompletionMessage {

	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: msg,
	}
}

func (a *Messages) MergeMessages(messages ...openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	return messages
}
