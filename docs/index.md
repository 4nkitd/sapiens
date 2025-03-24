# Sapiens Documentation

Welcome to the Sapiens documentation. Sapiens is a flexible Go framework for building AI agents with memory, tools, and embedding capabilities.

## Core Types

- [Agent](agent.md) - AI agent with configuration and capabilities
- [Prompt](prompt.md) - System prompt with versioning and enhancement
- [Tool](tool.md) - Capabilities that agents can use
- [Schema](schema.md) - JSON schema for defining data structures
- [Memory](memory.md) - Storage and retrieval system for agents
- [Embedding](embedding.md) - Vector embedding of text content
- [Vector](vector.md) - Numeric vector for embeddings
- [SimilarityResult](similarity-result.md) - Result of similarity search

## Getting Started

```go
package main

import (
	"fmt"
	"github.com/4nkitd/sapiens"
)

func main() {
	// Create a system prompt
	systemPrompt := sapiens.NewSystemPrompt(
		"You are a helpful assistant called {{.name}}.", 
		"v1.0",
	)
	
	// Create an agent
	agent := sapiens.Agent{
		Name:         "Sapiens Assistant",
		SystemPrompt: systemPrompt,
	}
	
	// Parse the prompt with parameters
	params := map[string]string{"name": "Sapiens"}
	parsedPrompt, err := systemPrompt.Parse(systemPrompt.Prompt, params)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Agent initialized with prompt:", parsedPrompt.Prompt)
}
```

For more detailed information, see the documentation for each specific type.
