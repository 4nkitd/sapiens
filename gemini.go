package main

import openai "github.com/sashabaranov/go-openai"

const (
	GeminiBaseUrl      = "https://generativelanguage.googleapis.com/v1beta/openai/"
	GeminiDefaultModel = "gemini-2.0-flash"
)

type GeminiInterface struct {
	BaseUrl      string
	DefaultModel string
	OrgId        string
	AuthToken    string
}

func NewGemini(authToken string) *GeminiInterface {
	instance_of_gemini := &GeminiInterface{
		BaseUrl:      GeminiBaseUrl,
		DefaultModel: GeminiDefaultModel,
		AuthToken:    authToken,
	}

	return instance_of_gemini

}

func (g *GeminiInterface) Client() *openai.Client {

	client_config := openai.DefaultConfig(g.AuthToken)

	client_config.BaseURL = g.BaseUrl

	client := openai.NewClientWithConfig(client_config)

	return client

}

func (g *GeminiInterface) GetDefaultModel() string {
	return g.DefaultModel
}
