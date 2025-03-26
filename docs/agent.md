# Agent

The `Agent` represents an AI agent with configuration, tools, and capabilities. It orchestrates interactions with language models, manages conversation history, and handles tool execution.

## Structure

```go
type Agent struct {
	Name                     string                 // Name of the agent
	LLM                      *LLM                   // Language model configuration
	SystemPrompts            []SystemPrompt         // System prompts that define the agent's behavior
	StructuredResponseSchema Schema                 // Schema for structured responses
	Tools                    []Tool                 // Available tools the agent can use
	toolImplementations      map[string]ToolImplementation // Tool implementations
	Messages                 []Message              // Current conversation messages
	conversationHistory      []Message              // Complete conversation history
	MaxRetry                 int                    // Maximum number of retry attempts
	Context                  map[string]interface{} // Context information
	MetaData                 map[string]interface{} // Metadata about the agent
	PromptManager            *PromptManager         // Manager for prompt templates
}
```

## Creation

```go
package main

import (
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create a new agent
	agent := sapiens.NewAgent(
		"MyAssistant",       // Name
		llm,                 // LLM implementation
		apiKey,              // API key
		"gemini-2.0-flash",  // Model name
		"google",            // Provider name (google or openai)
	)
}
```

## Key Methods

### System Prompts

System prompts define the agent's behavior and personality. See the [Prompt documentation](prompt.md) for more details.

```go
package main

import (
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider

	// Add a simple system prompt
	agent.AddSystemPrompt("You are a helpful assistant that provides concise answers.", "1.0")

	// Add a dynamic prompt from a template and data
	promptTemplate := `You are an assistant specialized in {{field}} who answers in {{style}} style.`
	cardData := map[string]interface{}{
		"field": "machine learning",
		"style": "clear and educational",
	}
	// agent.AddDynamicPrompt(promptTemplate, cardData, "1.1") // Removed AddDynamicPrompt

	// Create and manage reusable prompt templates
	// template := sapiens.PromptTemplate{ // Removed PromptTemplate
	// 	Name:        "expert_assistant",
	// 	Template:    `You are a {{level}} {{field}} expert. Help users with {{task}} questions.`,
	// 	Description: "Template for expert assistants in various fields",
	// 	Version:     "1.0",
	// }
	// agent.AddPromptTemplate(template)

	// Create a card from the template
	// card := sapiens.NewCard("expert_assistant", map[string]interface{}{ // Removed NewCard
	// 	"level": "senior",
	// 	"field": "software engineering",
	// 	"task":  "code review",
	// })

	// Apply the card template as a system prompt
	// agent.AddDynamicPromptWithCard(card, "1.0") // Removed AddDynamicPromptWithCard
}
```

### Managing Tools

Tools allow the agent to perform specific actions. See the [Tool documentation](tool.md) for more details.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider
	agent.AddSystemPrompt("You are a helpful assistant specialized in Go programming.", "1.0")

	// Add tools to the agent
	weatherTool := sapiens.Tool{
		Name:        "get_weather",
		Description: "Get current weather for a location",
		InputSchema: &sapiens.Schema{
			Type: "object",
			Properties: map[string]sapiens.Schema{
				"location": {Type: "string", Description: "City name"},
			},
			Required: []string{"location"},
		},
	}
	agent.AddTools(weatherTool)

	// Register tool implementation
	agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (string, error) {
		location, ok := params["location"].(string)
		if !ok {
			return "", fmt.Errorf("location must be a string")
		}
		// Implementation logic
		return fmt.Sprintf("The weather in %s is sunny and 25C", location), nil
	})

	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "What is the weather in London?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(response.Content)
}
```

### Context Management

Provide the agent with additional context to improve its responses. See the [LLM documentation](llm.md) for more details.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider
	agent.AddSystemPrompt("You are a helpful assistant specialized in Go programming.", "1.0")

	// Set context as a map
	agent.SetContext(map[string]interface{}{
		"user_id":     "user123",
		"preferences": map[string]string{"language": "English", "units": "metric"},
	})

	// Set context as a string
	agent.SetStringContext(`
User: John Smith
Subscription: Premium
Preferences: Dark mode, Technical responses
`)

	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "What are my preferences?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(response.Content)
}
```

### Structured Output

Request structured responses from the agent. See the [LLM documentation](llm.md) for more details.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider
	agent.AddSystemPrompt("You are a helpful assistant specialized in Go programming.", "1.0")

	// Define output schema
	schema := sapiens.Schema{
		Type: "object",
		Properties: map[string]sapiens.Schema{
			"summary":   {Type: "string", Description: "Brief summary"},
			"details":   {Type: "string", Description: "Detailed explanation"},
			"relevance": {Type: "number", Description: "Relevance score 0-10"},
		},
		Required: []string{"summary", "relevance"},
	}

	// Set schema for structured responses
	agent.SetStructuredResponseSchema(schema)

	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "Summarize the key features of the Go programming language.")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Access structured data if a schema was defined
	if response.Structured != nil {
		data := response.Structured.(map[string]interface{})
		fmt.Printf("Summary: %s (Relevance: %.1f)\n", data["summary"], data["relevance"])
	}
}
```

### Running the Agent

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4nkitd/sapiens"
	"github.com/4nkitd/sapiens/gemini" // or "github.com/4nkitd/sapiens/openai"
)

func main() {
	// Initialize the LLM client
	apiKey := os.Getenv("GEMINI_API_KEY") // or "OPENAI_API_KEY"
	llm := gemini.NewGoogleGenAI(apiKey, "gemini-2.0-flash") // or openai.NewOpenAI(apiKey, "gpt-4o")
	if err := llm.Initialize(); err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent("MyAssistant", llm, apiKey, "gemini-2.0-flash", "google") // or "openai" for provider
	agent.AddSystemPrompt("You are a helpful assistant specialized in Go programming.", "1.0")

	// Run the agent
	ctx := context.Background()
	response, err := agent.Run(ctx, "How do I implement a binary search tree in Go?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Access the response
	fmt.Println("Content:", response.Content)

	// Access tool calls if any were made
	for _, toolCall := range response.ToolCalls {
		fmt.Printf("Tool called: %s with inputs: %v\n", toolCall.Name, toolCall.InputMap)
	}

	// Access structured data if a schema was defined
	if response.Structured != nil {
		data := response.Structured.(map[string]interface{})
		fmt.Printf("Summary: %s (Relevance: %.1f)\n", data["summary"], data["relevance"])
	}

	// Access conversation history
	history := agent.GetHistory()
	fmt.Printf("Conversation has %d turns\n", len(history)/2)
}
```

## Related Types

- [LLM](llm.md)
- [SystemPrompt](prompt.md)
- [Tool](tool.md)
- [Schema](schema.md)
- [Message](message.md)
- [PromptManager](prompt.md#promptmanager)
