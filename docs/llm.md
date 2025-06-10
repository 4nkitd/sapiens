# LLM Providers

Sapiens supports multiple Language Model providers through a unified interface. All providers implement the same basic pattern: they return an OpenAI-compatible client that the Agent can use.

## Supported Providers

### OpenAI

OpenAI provides access to GPT models including GPT-4 and GPT-3.5.

```go
llm := NewOpenai(os.Getenv("OPENAI_API_KEY"))
```

**Configuration:**
- **Default Model:** `gpt-4.1-2025-04-14`
- **Base URL:** Uses OpenAI's default endpoint
- **Authentication:** API key via environment variable `OPENAI_API_KEY`

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    // Initialize OpenAI provider
    llm := NewOpenai(os.Getenv("OPENAI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Use the agent...
}
```

### Google Gemini

Google's Gemini models accessed through their OpenAI-compatible API.

```go
llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
```

**Configuration:**
- **Default Model:** `gemini-2.0-flash`
- **Base URL:** `https://generativelanguage.googleapis.com/v1beta/openai/`
- **Authentication:** API key via environment variable `GEMINI_API_KEY`

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    // Initialize Gemini provider
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Use the agent...
}
```

### Anthropic Claude

Anthropic's Claude models accessed through their API.

```go
llm := NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"))
```

**Configuration:**
- **Default Model:** `claude-sonet-3.5`
- **Base URL:** `https://generativelanguage.googleapis.com/v1beta/openai/`
- **Authentication:** API key via environment variable `ANTHROPIC_API_KEY`

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    // Initialize Anthropic provider
    llm := NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Use the agent...
}
```

### Ollama

Local models served through Ollama's OpenAI-compatible API.

```go
llm := NewOllama(baseUrl, authToken, modelName)
```

**Configuration:**
- **Default Model:** User-specified
- **Base URL:** User-specified (typically `http://localhost:11434/v1/`)
- **Authentication:** Optional auth token

**Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    // Initialize Ollama provider
    llm := NewOllama(
        "http://localhost:11434/v1/",  // Base URL
        "",                            // Auth token (optional for local)
        "llama2",                      // Model name
    )
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Use the agent...
}
```

## Provider Interface

All providers implement the same basic interface:

```go
type LLMProvider interface {
    Client() *openai.Client
    GetDefaultModel() string
}
```

### Client() Method

Returns an OpenAI-compatible client configured for the specific provider.

### GetDefaultModel() Method

Returns the default model name for the provider.

## Provider Implementation Details

### OpenAI Provider

```go
type OpenaiInterface struct {
    BaseUrl      string
    DefaultModel string
    OrgId        string
    AuthToken    string
}

func NewOpenai(authToken string) *OpenaiInterface
func (g *OpenaiInterface) Client() *openai.Client
func (g *OpenaiInterface) GetDefaultModel() string
```

### Gemini Provider

```go
type GeminiInterface struct {
    BaseUrl      string
    DefaultModel string
    OrgId        string
    AuthToken    string
}

func NewGemini(authToken string) *GeminiInterface
func (g *GeminiInterface) Client() *openai.Client
func (g *GeminiInterface) GetDefaultModel() string
```

### Anthropic Provider

```go
type AnthropicInterface struct {
    BaseUrl      string
    DefaultModel string
    OrgId        string
    AuthToken    string
}

func NewAnthropic(authToken string) *AnthropicInterface
func (g *AnthropicInterface) Client() *openai.Client
func (g *AnthropicInterface) GetDefaultModel() string
```

### Ollama Provider

```go
type OllamaInterface struct {
    BaseUrl      string
    DefaultModel string
    OrgId        string
    AuthToken    string
}

func NewOllama(baseUrl, authToken, defaultModel string) *OllamaInterface
func (g *OllamaInterface) Client() *openai.Client
func (g *OllamaInterface) GetDefaultModel() string
```

## Switching Between Providers

You can easily switch between providers by changing the LLM initialization:

```go
package main

import (
    "context"
    "os"
)

func main() {
    var llm LLMProvider
    
    // Choose provider based on environment or configuration
    provider := os.Getenv("LLM_PROVIDER")
    switch provider {
    case "openai":
        llm = NewOpenai(os.Getenv("OPENAI_API_KEY"))
    case "gemini":
        llm = NewGemini(os.Getenv("GEMINI_API_KEY"))
    case "anthropic":
        llm = NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"))
    case "ollama":
        llm = NewOllama("http://localhost:11434/v1/", "", "llama2")
    default:
        llm = NewGemini(os.Getenv("GEMINI_API_KEY"))
    }
    
    // Create agent with chosen provider
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant",
    )
    
    // Use the agent normally...
}
```

## Environment Setup

Set the appropriate environment variables for your chosen providers:

```bash
# OpenAI
export OPENAI_API_KEY="sk-your-openai-key"

# Google Gemini
export GEMINI_API_KEY="your-gemini-api-key"

# Anthropic
export ANTHROPIC_API_KEY="your-anthropic-key"

# Optional: Choose default provider
export LLM_PROVIDER="gemini"
```

## Error Handling

All providers use the same error handling patterns through the OpenAI client:

```go
resp, err := agent.Ask(messages)
if err != nil {
    // Handle provider-specific errors
    switch {
    case strings.Contains(err.Error(), "authentication"):
        log.Printf("Authentication error - check your API key")
    case strings.Contains(err.Error(), "rate limit"):
        log.Printf("Rate limit exceeded - try again later")
    case strings.Contains(err.Error(), "model"):
        log.Printf("Model error - check model name and availability")
    default:
        log.Printf("Provider error: %v", err)
    }
    return
}
```

## Model Selection

While each provider has a default model, you can specify a different model when creating the agent:

```go
// Use default model
agent := NewAgent(ctx, llm.Client(), llm.GetDefaultModel(), systemPrompt)

// Use specific model
agent := NewAgent(ctx, llm.Client(), "gpt-4", systemPrompt)
agent := NewAgent(ctx, llm.Client(), "gemini-1.5-pro", systemPrompt)
agent := NewAgent(ctx, llm.Client(), "claude-3-opus", systemPrompt)
```

## Performance Considerations

### OpenAI
- Generally fast response times
- Rate limits based on subscription tier
- Higher cost for advanced models like GPT-4

### Gemini
- Fast response times
- Generous free tier
- Good performance for most tasks

### Anthropic
- High-quality responses
- Good for complex reasoning tasks
- Rate limits based on subscription

### Ollama
- No API costs (runs locally)
- Response time depends on hardware
- Full privacy and control
- Requires local setup and model downloads

## Related Documentation

- [Agent](agent.md) - Using providers with agents
- [Tools](tool.md) - Tool compatibility across providers
- [Examples](examples.md) - Provider-specific examples