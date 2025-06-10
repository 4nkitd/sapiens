package sapiens

import openai "github.com/sashabaranov/go-openai"

const (
	OpenaiDefaultModel = "gpt-4.1-2025-04-14"
)

type OpenaiInterface struct {
	BaseUrl      string
	DefaultModel string
	OrgId        string
	AuthToken    string
}

func NewOpenai(authToken string) *OpenaiInterface {
	instance_of_openai := &OpenaiInterface{
		DefaultModel: OpenaiDefaultModel,
		AuthToken:    authToken,
	}

	return instance_of_openai

}

func (g *OpenaiInterface) Client() *openai.Client {

	client_config := openai.DefaultConfig(g.AuthToken)

	client_config.BaseURL = g.BaseUrl

	client := openai.NewClientWithConfig(client_config)

	return client

}

func (g *OpenaiInterface) GetDefaultModel() string {
	return g.DefaultModel
}
