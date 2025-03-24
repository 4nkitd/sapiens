# Sapiens: Go AI Agent Framework

Sapiens is a lightweight, flexible framework for building AI agents in Go. It provides tools for creating, managing, and deploying intelligent agents with memory, embedding capabilities, and tool integration.

## Features

- **Agent Management**: Create and manage AI agents with customizable system prompts
- **Memory Systems**: In-memory storage with embedding-based retrieval
- **Tool Integration**: Define and use tools with input/output schemas
- **Prompt Templates**: Dynamic prompt generation with templating support
- **Vector Embeddings**: Create and compare text embeddings

## Installation

```bash
go get github.com/4nkitd/sapiens
```

## Quick Start

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
	
	// Create memory for the agent
	memory := sapiens.NewMemory("semantic", map[string]interface{}{
		"dimension": 1536,
	})
	
	// Add the memory to the agent
	agent.Memory = append(agent.Memory, memory)
	
	fmt.Println("Agent initialized with prompt:", parsedPrompt.Prompt)
}
```

## Core Components

### Agent

The base entity that combines a system prompt, tools, memory, and other components.

```go
agent := sapiens.Agent{
    Name: "My Assistant",
    SystemPrompt: systemPrompt,
    Tools: []Tool{},
    Memory: []Memory{},
    MaxRetry: 3,
}
```

### Memory

Used for storing and retrieving information, with support for vector-based semantic search.

```go
memory := sapiens.NewMemory("semantic", map[string]interface{}{
    "dimension": 1536,
})

// Add an embedding to memory
embedding := sapiens.NewEmbedding(agent, "This is important information")
memory.Add("info1", data, embedding)

// Search similar embeddings
results := memory.Search(queryEmbedding)
```

### Tools

Define tools that your agent can use to accomplish tasks.

```go
tool := sapiens.NewTool(
    "weather_lookup", 
    "Look up the current weather for a location", 
    []sapiens.Tool{},
)

// Define input schema
tool.AddInputSchema(&sapiens.Schema{
    Type: "object",
    Properties: map[string]sapiens.Schema{
        "location": {Type: "string"},
    },
    Required: []string{"location"},
})
```

### Prompts

Create and manage system prompts with templating support.

```go
prompt := sapiens.NewSystemPrompt(
    "You are {{.role}} focused on helping with {{.domain}}.",
    "v1.0",
)

params := map[string]string{
    "role": "an expert assistant",
    "domain": "programming",
}

parsedPrompt, err := prompt.Parse(prompt.Prompt, params)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
