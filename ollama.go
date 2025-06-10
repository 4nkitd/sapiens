package main

import openai "github.com/sashabaranov/go-openai"

const (
	OllamaBaseUrl      = ""
	OllamaDefaultModel = ""
)

type OllamaInterface struct {
	BaseUrl      string
	DefaultModel string
	OrgId        string
	AuthToken    string
}

func NewOllama(baseUrl, authToken, defaultModel string) *OllamaInterface {
	instance_of_gemini := &OllamaInterface{
		BaseUrl:      baseUrl,
		DefaultModel: defaultModel,
		AuthToken:    authToken,
	}

	return instance_of_gemini

}

func (g *OllamaInterface) Client() *openai.Client {

	client_config := openai.DefaultConfig(g.AuthToken)

	client_config.BaseURL = g.BaseUrl

	client := openai.NewClientWithConfig(client_config)

	return client

}

func (g *OllamaInterface) GetDefaultModel() string {
	return g.DefaultModel
}
