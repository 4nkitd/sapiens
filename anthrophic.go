package sapiens

import openai "github.com/sashabaranov/go-openai"

const (
	AnthropicBaseUrl      = "https://generativelanguage.googleapis.com/v1beta/openai/"
	AnthropicDefaultModel = "claude-sonet-3.5"
)

type AnthropicInterface struct {
	BaseUrl      string
	DefaultModel string
	OrgId        string
	AuthToken    string
}

func NewAnthropic(authToken string) *AnthropicInterface {
	instance_of_gemini := &AnthropicInterface{
		BaseUrl:      AnthropicBaseUrl,
		DefaultModel: AnthropicDefaultModel,
		AuthToken:    authToken,
	}

	return instance_of_gemini

}

func (g *AnthropicInterface) Client() *openai.Client {

	client_config := openai.DefaultConfig(g.AuthToken)

	client_config.BaseURL = g.BaseUrl

	client := openai.NewClientWithConfig(client_config)

	return client

}

func (g *AnthropicInterface) GetDefaultModel() string {
	return g.DefaultModel
}
