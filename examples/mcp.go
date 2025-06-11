package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
)

func main() {

	client := sapiens.NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := sapiens.NewAgent(
		context.Background(),
		client.Client(),
		"gemini-2.5-pro-preview-06-05",
		"You are a helpful assistant with access to both regular tools and MCP tools.",
	)

	// Add a regular tool
	// err := agent.AddTool(
	// 	"get_weather",
	// 	"Get weather information for a specific location",
	// 	map[string]jsonschema.Definition{
	// 		"location": {
	// 			Type:        jsonschema.String,
	// 			Description: "City name or timezone (e.g., New York, Tokyo, UTC)",
	// 		},
	// 	},
	// 	[]string{"location"},
	// 	func(params map[string]string) string {
	// 		location := params["location"]
	// 		return fmt.Sprintf("Weather in %s: 72°F, sunny", location)
	// 	},
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to add regular tool: %v", err)
	// }

	// Try to add MCP server (optional - will continue without it if not available)
	mcpURL := "http://localhost:8080/sse"
	err := agent.AddMCP(mcpURL, nil)
	if err != nil {
		fmt.Printf("Warning: Could not connect to MCP server at %s: %v\n", mcpURL, err)
		fmt.Println("Continuing with regular tools only...")
		fmt.Println("To test MCP functionality, please see MCP_SETUP.md for server setup instructions.")
	} else {
		fmt.Printf("Successfully connected to MCP server at %s\n", mcpURL)
	}

	fmt.Printf("Agent has %d regular tools and %d MCP tools\n", len(agent.Tools), len(agent.McpTools))

	// Ask a question that uses regular tools and potentially MCP tools
	// var question string
	// if len(agent.McpTools) > 0 {
	// 	question = "What's the weather like in New York? Also, can you use any MCP tools to help me?"
	// } else {
	// 	question = "What's the weather like in New York? Please use the available weather tool."
	// }

	question := "create link 200 rupee from monu vis a@paytring.com email"

	response, err := agent.Ask(sapiens.NewMessages().MergeMessages(
		sapiens.NewMessages().UserMessage(question)))
	if err != nil {
		log.Fatalf("Failed to get response from agent: %v", err)
	}

	if len(response.Choices) > 0 {
		fmt.Println("\nAgent response:")
		fmt.Println(response.Choices[0].Message.Content)
	} else {
		fmt.Println("No response received from agent")
	}

}
