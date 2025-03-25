# Sapiens Examples

This document provides practical examples of how to use the Sapiens library for different use cases.

## Basic Agent Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
)

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent(llm)

	// Execute a simple prompt
	ctx := context.Background()
	result, err := agent.Execute(ctx, "What are the three laws of robotics?")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}

	fmt.Println(result)
}
```

## Using System Prompts

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
)

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent with system prompt
	agent := sapiens.NewAgent(
		llm,
		sapiens.WithSystemPrompt("You are a helpful coding assistant specialized in Go programming."),
	)

	// Execute a coding-related prompt
	ctx := context.Background()
	result, err := agent.Execute(ctx, "Write a function to find the fibonacci number at a given position.")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}

	fmt.Println(result)
}
```

## Streaming Responses

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
)

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent(llm)

	// Execute with streaming
	ctx := context.Background()
	responseStream, errStream := agent.ExecuteWithStream(ctx, "Write a short story about an AI that becomes sentient.")

	// Process the streaming response
	for {
		select {
		case chunk, ok := <-responseStream:
			if !ok {
				responseStream = nil
				if errStream == nil {
					return
				}
				continue
			}
			fmt.Print(chunk) // Print each chunk as it arrives
		case err, ok := <-errStream:
			if !ok {
				errStream = nil
				if responseStream == nil {
					return
				}
				continue
			}
			if err != nil {
				log.Fatalf("Error in stream: %v", err)
			}
		}
	}
}
```

## Function Calling with Tools

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
	"github.com/yourusername/sapiens/tools"
)

// Define the functions to be used as tools
func getWeather(location string) string {
	// In a real application, this would call a weather API
	return fmt.Sprintf("It's sunny and 72 degrees in %s.", location)
}

func calculateTip(amount float64, percentage float64) string {
	tip := amount * (percentage / 100)
	total := amount + tip
	return fmt.Sprintf("Tip: $%.2f, Total: $%.2f", tip, total)
}

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent(llm)

	// Create and register tools
	weatherTool := tools.NewTool("getWeather", "Get the current weather for a location", getWeather)
	tipTool := tools.NewTool("calculateTip", "Calculate tip amount and total bill", calculateTip)
	agent.RegisterTool(weatherTool)
	agent.RegisterTool(tipTool)

	// Test the tools
	ctx := context.Background()
	
	// Test weather tool
	weatherResponse, err := agent.Execute(ctx, "What's the weather like in San Francisco?")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}
	fmt.Println(weatherResponse)
	
	// Test tip calculator
	tipResponse, err := agent.Execute(ctx, "I need to calculate a 15% tip on a $75.50 bill.")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}
	fmt.Println(tipResponse)
}
```

## Multi-turn Conversations

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
)

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent(llm)
	ctx := context.Background()

	// First turn
	resp1, err := agent.Execute(ctx, "My name is Alex and I'm learning Go programming.")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}
	fmt.Println("AI:", resp1)

	// Second turn (should remember the context)
	resp2, err := agent.Execute(ctx, "What's a good project for me to practice with?")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}
	fmt.Println("AI:", resp2)

	// Third turn (continuing the conversation)
	resp3, err := agent.Execute(ctx, "Can you show me some sample code for that project?")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}
	fmt.Println("AI:", resp3)

	// Access conversation history
	history := agent.GetHistory()
	fmt.Printf("\nConversation had %d turns\n", len(history)/2) // Divide by 2 because each turn has user + assistant message
}
```

## Using with Anthropic Claude

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/anthropic"
)

func main() {
	// Initialize the Anthropic LLM client
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-opus-20240229"),
	)
	if err != nil {
		log.Fatalf("Error initializing Anthropic client: %v", err)
	}

	// Create an agent with Claude
	agent := sapiens.NewAgent(llm)

	// Execute a prompt
	ctx := context.Background()
	result, err := agent.Execute(ctx, "Explain how quantum computing differs from classical computing.")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}

	fmt.Println(result)
}
```

## Advanced Tool Usage (Web Browsing)

```go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/yourusername/sapiens"
	"github.com/yourusername/sapiens/llm/openai"
	"github.com/yourusername/sapiens/tools"
)

// Simple web fetching function
func fetchWebContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// Return the first 1000 chars to avoid overwhelming the context
	content := string(body)
	if len(content) > 1000 {
		content = content[:1000] + "... [content truncated]"
	}
	
	return content, nil
}

func main() {
	// Initialize the LLM client
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error initializing OpenAI client: %v", err)
	}

	// Create an agent
	agent := sapiens.NewAgent(
		llm,
		sapiens.WithSystemPrompt("You are a helpful research assistant that can browse the web."),
	)

	// Create and register the web fetching tool
	webTool := tools.NewTool(
		"fetchWebContent",
		"Fetch content from a web page URL",
		func(url string) string {
			content, err := fetchWebContent(url)
			if err != nil {
				return fmt.Sprintf("Error fetching URL: %v", err)
			}
			return content
		},
	)
	agent.RegisterTool(webTool)

	// Test the web browsing capability
	ctx := context.Background()
	result, err := agent.Execute(ctx, "What's on the front page of example.com?")
	if err != nil {
		log.Fatalf("Error executing agent: %v", err)
	}

	fmt.Println(result)
}
```
