# Examples

This document provides practical examples of using Sapiens in real-world scenarios. Each example demonstrates different features and usage patterns.

## Table of Contents

1. [Basic Agent Setup](#basic-agent-setup)
2. [Weather Assistant](#weather-assistant)
3. [Task Management System](#task-management-system)
4. [Multi-Tool Agent](#multi-tool-agent)
5. [Structured Output Examples](#structured-output-examples)
6. [Provider Switching](#provider-switching)
7. [Error Handling Patterns](#error-handling-patterns)
8. [Advanced Use Cases](#advanced-use-cases)

## Basic Agent Setup

### Simple Question and Answer

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    // Initialize provider
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Create agent
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful programming assistant",
    )
    
    // Create messages
    message := NewMessages()
    
    // Ask a question
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage("Explain the difference between arrays and slices in Go"),
    ))
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Println("Response:", resp.Choices[0].Message.Content)
}
```

### Conversation with Context

```go
func conversationExample() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant. Remember our conversation context.",
    )
    
    message := NewMessages()
    
    // First interaction
    resp1, _ := agent.Ask(message.MergeMessages(
        message.UserMessage("I'm building a REST API in Go. What framework should I use?"),
    ))
    fmt.Println("First response:", resp1.Choices[0].Message.Content)
    
    // Second interaction (builds on previous context)
    resp2, _ := agent.Ask(message.MergeMessages(
        message.UserMessage("How do I handle authentication in that framework?"),
    ))
    fmt.Println("Second response:", resp2.Choices[0].Message.Content)
}
```

## Weather Assistant

A complete weather assistant with error handling and multiple features.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

type WeatherData struct {
    Temperature string `json:"temperature"`
    Condition   string `json:"condition"`
    Humidity    string `json:"humidity"`
    WindSpeed   string `json:"wind_speed"`
    Location    string `json:"location"`
    Timestamp   string `json:"timestamp"`
}

func weatherTool(params map[string]string) string {
    location := params["location"]
    unit := params["unit"]
    if unit == "" {
        unit = "celsius"
    }
    
    // Simulate API call (in real implementation, call actual weather API)
    var temp string
    switch location {
    case "London":
        temp = "15"
        if unit == "fahrenheit" {
            temp = "59"
        }
    case "New York":
        temp = "20"
        if unit == "fahrenheit" {
            temp = "68"
        }
    case "Tokyo":
        temp = "25"
        if unit == "fahrenheit" {
            temp = "77"
        }
    default:
        temp = "22"
        if unit == "fahrenheit" {
            temp = "72"
        }
    }
    
    weather := WeatherData{
        Temperature: temp + "°" + map[string]string{"celsius": "C", "fahrenheit": "F"}[unit],
        Condition:   "Partly cloudy",
        Humidity:    "65%",
        WindSpeed:   "10 km/h",
        Location:    location,
        Timestamp:   time.Now().Format(time.RFC3339),
    }
    
    result, _ := json.Marshal(weather)
    return string(result)
}

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a weather assistant. Provide helpful weather information and advice.",
    )
    
    // Add weather tool
    agent.AddTool(
        "get_weather",
        "Get current weather information for any location",
        map[string]jsonschema.Definition{
            "location": {
                Type:        jsonschema.String,
                Description: "City name (e.g., 'London', 'New York', 'Tokyo')",
            },
            "unit": {
                Type:        jsonschema.String,
                Enum:        []string{"celsius", "fahrenheit"},
                Description: "Temperature unit preference",
            },
        },
        []string{"location"},
        weatherTool,
    )
    
    message := NewMessages()
    
    // Example queries
    queries := []string{
        "What's the weather like in London?",
        "How's the weather in New York? Use Fahrenheit please.",
        "Compare the weather between Tokyo and London",
        "Should I bring an umbrella if I'm going to London today?",
    }
    
    for _, query := range queries {
        fmt.Printf("\n=== Query: %s ===\n", query)
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(query),
        ))
        
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Println("Response:", resp.Choices[0].Message.Content)
    }
}
```

## Task Management System

A comprehensive task management system with multiple tools and structured responses.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

type Task struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Priority    string   `json:"priority"`
    Status      string   `json:"status"`
    DueDate     string   `json:"due_date"`
    Tags        []string `json:"tags"`
    CreatedAt   string   `json:"created_at"`
}

type TaskResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Task    *Task  `json:"task,omitempty"`
    Tasks   []Task `json:"tasks,omitempty"`
}

var tasks []Task
var taskCounter = 1

func createTaskTool(params map[string]string) string {
    task := Task{
        ID:          "task_" + strconv.Itoa(taskCounter),
        Title:       params["title"],
        Description: params["description"],
        Priority:    params["priority"],
        Status:      "pending",
        DueDate:     params["due_date"],
        Tags:        []string{},
        CreatedAt:   time.Now().Format(time.RFC3339),
    }
    
    if task.Priority == "" {
        task.Priority = "medium"
    }
    
    tasks = append(tasks, task)
    taskCounter++
    
    response := TaskResponse{
        Success: true,
        Message: "Task created successfully",
        Task:    &task,
    }
    
    result, _ := json.Marshal(response)
    return string(result)
}

func listTasksTool(params map[string]string) string {
    status := params["status"]
    priority := params["priority"]
    
    filteredTasks := []Task{}
    for _, task := range tasks {
        if status != "" && task.Status != status {
            continue
        }
        if priority != "" && task.Priority != priority {
            continue
        }
        filteredTasks = append(filteredTasks, task)
    }
    
    response := TaskResponse{
        Success: true,
        Message: fmt.Sprintf("Found %d tasks", len(filteredTasks)),
        Tasks:   filteredTasks,
    }
    
    result, _ := json.Marshal(response)
    return string(result)
}

func updateTaskTool(params map[string]string) string {
    taskID := params["task_id"]
    newStatus := params["status"]
    
    for i, task := range tasks {
        if task.ID == taskID {
            if newStatus != "" {
                tasks[i].Status = newStatus
            }
            
            response := TaskResponse{
                Success: true,
                Message: "Task updated successfully",
                Task:    &tasks[i],
            }
            
            result, _ := json.Marshal(response)
            return string(result)
        }
    }
    
    response := TaskResponse{
        Success: false,
        Message: "Task not found",
    }
    
    result, _ := json.Marshal(response)
    return string(result)
}

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a task management assistant. Help users create, manage, and track their tasks efficiently.",
    )
    
    // Add create task tool
    agent.AddTool(
        "create_task",
        "Create a new task",
        map[string]jsonschema.Definition{
            "title": {
                Type:        jsonschema.String,
                Description: "Task title",
            },
            "description": {
                Type:        jsonschema.String,
                Description: "Detailed task description",
            },
            "priority": {
                Type:        jsonschema.String,
                Enum:        []string{"low", "medium", "high", "urgent"},
                Description: "Task priority level",
            },
            "due_date": {
                Type:        jsonschema.String,
                Description: "Due date in YYYY-MM-DD format",
            },
        },
        []string{"title"},
        createTaskTool,
    )
    
    // Add list tasks tool
    agent.AddTool(
        "list_tasks",
        "List tasks with optional filtering",
        map[string]jsonschema.Definition{
            "status": {
                Type:        jsonschema.String,
                Enum:        []string{"pending", "in_progress", "completed", "cancelled"},
                Description: "Filter by task status",
            },
            "priority": {
                Type:        jsonschema.String,
                Enum:        []string{"low", "medium", "high", "urgent"},
                Description: "Filter by priority level",
            },
        },
        []string{},
        listTasksTool,
    )
    
    // Add update task tool
    agent.AddTool(
        "update_task",
        "Update an existing task",
        map[string]jsonschema.Definition{
            "task_id": {
                Type:        jsonschema.String,
                Description: "ID of the task to update",
            },
            "status": {
                Type:        jsonschema.String,
                Enum:        []string{"pending", "in_progress", "completed", "cancelled"},
                Description: "New status for the task",
            },
        },
        []string{"task_id"},
        updateTaskTool,
    )
    
    message := NewMessages()
    
    // Interactive task management session
    commands := []string{
        "Create a high priority task to 'Review project documentation' due tomorrow",
        "Create a task to 'Update website design' with medium priority",
        "List all my tasks",
        "Mark the documentation review task as completed",
        "Show me only the pending tasks",
    }
    
    for _, command := range commands {
        fmt.Printf("\n=== User: %s ===\n", command)
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(command),
        ))
        
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Println("Assistant:", resp.Choices[0].Message.Content)
    }
}
```

## Multi-Tool Agent

An agent that can handle multiple different types of requests with various tools.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "os"
    "strconv"
    "strings"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

func calculatorTool(params map[string]string) string {
    expression := params["expression"]
    
    // Simple calculator (in production, use a proper expression parser)
    expression = strings.ReplaceAll(expression, " ", "")
    
    var result float64
    var err error
    
    if strings.Contains(expression, "+") {
        parts := strings.Split(expression, "+")
        if len(parts) == 2 {
            a, _ := strconv.ParseFloat(parts[0], 64)
            b, _ := strconv.ParseFloat(parts[1], 64)
            result = a + b
        }
    } else if strings.Contains(expression, "*") {
        parts := strings.Split(expression, "*")
        if len(parts) == 2 {
            a, _ := strconv.ParseFloat(parts[0], 64)
            b, _ := strconv.ParseFloat(parts[1], 64)
            result = a * b
        }
    } else {
        result, err = strconv.ParseFloat(expression, 64)
        if err != nil {
            return `{"error": "Invalid expression", "result": null}`
        }
    }
    
    return fmt.Sprintf(`{"expression": "%s", "result": %.2f, "success": true}`, expression, result)
}

func currencyConverterTool(params map[string]string) string {
    amount, _ := strconv.ParseFloat(params["amount"], 64)
    from := strings.ToUpper(params["from"])
    to := strings.ToUpper(params["to"])
    
    // Mock exchange rates
    rates := map[string]float64{
        "USD": 1.0,
        "EUR": 0.85,
        "GBP": 0.73,
        "JPY": 110.0,
        "CAD": 1.25,
    }
    
    fromRate, fromExists := rates[from]
    toRate, toExists := rates[to]
    
    if !fromExists || !toExists {
        return `{"error": "Unsupported currency", "result": null}`
    }
    
    // Convert to USD first, then to target currency
    usdAmount := amount / fromRate
    convertedAmount := usdAmount * toRate
    
    return fmt.Sprintf(`{
        "original_amount": %.2f,
        "original_currency": "%s",
        "converted_amount": %.2f,
        "target_currency": "%s",
        "exchange_rate": %.4f,
        "success": true
    }`, amount, from, convertedAmount, to, toRate/fromRate)
}

func timeTool(params map[string]string) string {
    timezone := params["timezone"]
    if timezone == "" {
        timezone = "UTC"
    }
    
    now := time.Now()
    
    // Mock timezone handling (in production, use proper timezone library)
    var timeStr string
    switch strings.ToUpper(timezone) {
    case "EST", "EASTERN":
        timeStr = now.Add(-5 * time.Hour).Format("15:04:05 MST")
    case "PST", "PACIFIC":
        timeStr = now.Add(-8 * time.Hour).Format("15:04:05 MST")
    case "JST", "JAPAN":
        timeStr = now.Add(9 * time.Hour).Format("15:04:05 MST")
    default:
        timeStr = now.UTC().Format("15:04:05 UTC")
    }
    
    return fmt.Sprintf(`{
        "current_time": "%s",
        "timezone": "%s",
        "timestamp": "%s",
        "success": true
    }`, timeStr, timezone, now.Format(time.RFC3339))
}

func unitConverterTool(params map[string]string) string {
    value, _ := strconv.ParseFloat(params["value"], 64)
    from := strings.ToLower(params["from"])
    to := strings.ToLower(params["to"])
    
    // Length conversions (to meters)
    lengthToMeters := map[string]float64{
        "m": 1.0, "meter": 1.0, "meters": 1.0,
        "cm": 0.01, "centimeter": 0.01, "centimeters": 0.01,
        "km": 1000.0, "kilometer": 1000.0, "kilometers": 1000.0,
        "ft": 0.3048, "foot": 0.3048, "feet": 0.3048,
        "in": 0.0254, "inch": 0.0254, "inches": 0.0254,
    }
    
    fromFactor, fromExists := lengthToMeters[from]
    toFactor, toExists := lengthToMeters[to]
    
    if !fromExists || !toExists {
        return `{"error": "Unsupported unit conversion", "result": null}`
    }
    
    // Convert to meters, then to target unit
    meters := value * fromFactor
    result := meters / toFactor
    
    return fmt.Sprintf(`{
        "original_value": %.2f,
        "original_unit": "%s",
        "converted_value": %.4f,
        "target_unit": "%s",
        "success": true
    }`, value, from, result, to)
}

func main() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a multi-purpose assistant with access to calculator, currency converter, time, and unit conversion tools. Help users with various calculations and conversions.",
    )
    
    // Add calculator tool
    agent.AddTool(
        "calculate",
        "Perform mathematical calculations",
        map[string]jsonschema.Definition{
            "expression": {
                Type:        jsonschema.String,
                Description: "Mathematical expression (e.g., '15+25', '100*0.8')",
            },
        },
        []string{"expression"},
        calculatorTool,
    )
    
    // Add currency converter
    agent.AddTool(
        "convert_currency",
        "Convert between different currencies",
        map[string]jsonschema.Definition{
            "amount": {
                Type:        jsonschema.String,
                Description: "Amount to convert",
            },
            "from": {
                Type:        jsonschema.String,
                Description: "Source currency code (USD, EUR, GBP, JPY, CAD)",
            },
            "to": {
                Type:        jsonschema.String,
                Description: "Target currency code (USD, EUR, GBP, JPY, CAD)",
            },
        },
        []string{"amount", "from", "to"},
        currencyConverterTool,
    )
    
    // Add time tool
    agent.AddTool(
        "get_time",
        "Get current time in different timezones",
        map[string]jsonschema.Definition{
            "timezone": {
                Type:        jsonschema.String,
                Description: "Timezone (UTC, EST, PST, JST, etc.)",
            },
        },
        []string{},
        timeTool,
    )
    
    // Add unit converter
    agent.AddTool(
        "convert_units",
        "Convert between different units of measurement",
        map[string]jsonschema.Definition{
            "value": {
                Type:        jsonschema.String,
                Description: "Value to convert",
            },
            "from": {
                Type:        jsonschema.String,
                Description: "Source unit (m, cm, km, ft, in)",
            },
            "to": {
                Type:        jsonschema.String,
                Description: "Target unit (m, cm, km, ft, in)",
            },
        },
        []string{"value", "from", "to"},
        unitConverterTool,
    )
    
    message := NewMessages()
    
    // Test various capabilities
    requests := []string{
        "What is 15 * 37 + 125?",
        "Convert 1000 USD to EUR",
        "What time is it in Tokyo?",
        "Convert 5 feet to meters",
        "I have 500 EUR, how much is that in Japanese Yen? Also what time is it in Japan?",
        "Calculate 25% of 200, then convert that amount from USD to GBP",
    }
    
    for i, request := range requests {
        fmt.Printf("\n=== Example %d ===\n", i+1)
        fmt.Printf("User: %s\n", request)
        
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(request),
        ))
        
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Printf("Assistant: %s\n", resp.Choices[0].Message.Content)
    }
}
```

## Structured Output Examples

Examples of using structured responses for different scenarios.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

// Product Analysis Response
type ProductAnalysis struct {
    ProductName  string              `json:"product_name"`
    Summary      string              `json:"summary"`
    Pros         []string            `json:"pros"`
    Cons         []string            `json:"cons"`
    Rating       float64             `json:"rating"`
    Price        string              `json:"price"`
    Competitors  []Competitor        `json:"competitors"`
    Recommendation string            `json:"recommendation"`
}

type Competitor struct {
    Name   string  `json:"name"`
    Price  string  `json:"price"`
    Rating float64 `json:"rating"`
}

// Recipe Response
type Recipe struct {
    Name         string       `json:"name"`
    Description  string       `json:"description"`
    PrepTime     string       `json:"prep_time"`
    CookTime     string       `json:"cook_time"`
    Servings     int          `json:"servings"`
    Difficulty   string       `json:"difficulty"`
    Ingredients  []Ingredient `json:"ingredients"`
    Instructions []string     `json:"instructions"`
    Tips         []string     `json:"tips"`
}

type Ingredient struct {
    Name     string `json:"name"`
    Amount   string `json:"amount"`
    Unit     string `json:"unit"`
    Optional bool   `json:"optional"`
}

// Code Review Response
type CodeReview struct {
    OverallScore int           `json:"overall_score"`
    Summary      string        `json:"summary"`
    Issues       []CodeIssue   `json:"issues"`
    Suggestions  []Suggestion  `json:"suggestions"`
    Strengths    []string      `json:"strengths"`
    NextSteps    []string      `json:"next_steps"`
}

type CodeIssue struct {
    Line        int    `json:"line"`
    Severity    string `json:"severity"`
    Type        string `json:"type"`
    Description string `json:"description"`
    Solution    string `json:"solution"`
}

type Suggestion struct {
    Category    string `json:"category"`
    Description string `json:"description"`
    Impact      string `json:"impact"`
}

func structuredOutputExamples() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    
    // Example 1: Product Analysis
    fmt.Println("=== Product Analysis Example ===")
    agent1 := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a product analyst. Provide detailed, structured analysis of products.",
    )
    
    var productAnalysis ProductAnalysis
    agent1.SetResponseSchema("product_analysis", "Structured product analysis", true, productAnalysis)
    
    message := NewMessages()
    resp1, err := agent1.Ask(message.MergeMessages(
        message.UserMessage("Analyze the iPhone 15 Pro. Include pros, cons, rating, and competitors."),
    ))
    
    if err == nil {
        err = agent1.ParseResponse(resp1, &productAnalysis)
        if err == nil {
            fmt.Printf("Product: %s\n", productAnalysis.ProductName)
            fmt.Printf("Rating: %.1f/10\n", productAnalysis.Rating)
            fmt.Printf("Summary: %s\n", productAnalysis.Summary)
            fmt.Printf("Pros: %v\n", productAnalysis.Pros)
            fmt.Printf("Cons: %v\n", productAnalysis.Cons)
        }
    }
    
    // Example 2: Recipe Generation
    fmt.Println("\n=== Recipe Example ===")
    agent2 := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a professional chef. Create detailed recipes with structured ingredients and instructions.",
    )
    
    var recipe Recipe
    agent2.SetResponseSchema("recipe", "Structured recipe with ingredients and instructions", true, recipe)
    
    resp2, err := agent2.Ask(message.MergeMessages(
        message.UserMessage("Create a recipe for chocolate chip cookies that serves 4 people."),
    ))
    
    if err == nil {
        err = agent2.ParseResponse(resp2, &recipe)
        if err == nil {
            fmt.Printf("Recipe: %s\n", recipe.Name)
            fmt.Printf("Prep Time: %s, Cook Time: %s\n", recipe.PrepTime, recipe.CookTime)
            fmt.Printf("Servings: %d, Difficulty: %s\n", recipe.Servings, recipe.Difficulty)
            fmt.Println("Ingredients:")
            for _, ing := range recipe.Ingredients {
                optional := ""
                if ing.Optional {
                    optional = " (optional)"
                }
                fmt.Printf("  - %s %s %s%s\n", ing.Amount, ing.Unit, ing.Name, optional)
            }
            fmt.Println("Instructions:")
            for i, inst := range recipe.Instructions {
                fmt.Printf("  %d. %s\n", i+1, inst)
            }
        }
    }
    
    // Example 3: Code Review
    fmt.Println("\n=== Code Review Example ===")
    agent3 := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a senior software engineer. Provide detailed code reviews with structured feedback.",
    )
    
    var codeReview CodeReview
    agent3.SetResponseSchema("code_review", "Structured code review with issues and suggestions", true, codeReview)
    
    codeToReview := `
func processUser(user *User) error {
    if user == nil {
        return errors.New("user is nil")
    }
    
    // Validate email
    if user.Email == "" {
        return errors.New("email required")
    }
    
    // Save to database
    db.Save(user)
    
    return nil
}
`
    
    resp3, err := agent3.Ask(message.MergeMessages(
        message.UserMessage("Review this Go code and provide structured feedback:\n" + codeToReview),
    ))
    
    if err == nil {
        err = agent3.ParseResponse(resp3, &codeReview)
        if err == nil {
            fmt.Printf("Overall Score: %d/100\n", codeReview.OverallScore)
            fmt.Printf("Summary: %s\n", codeReview.Summary)
            fmt.Println("Issues:")
            for _, issue := range codeReview.Issues {
                fmt.Printf("  Line %d [%s]: %s\n", issue.Line, issue.Severity, issue.Description)
                fmt.Printf("    Solution: %s\n", issue.Solution)
            }
            fmt.Println("Suggestions:")
            for _, suggestion := range codeReview.Suggestions {
                fmt.Printf("  [%s] %s (Impact: %s)\n", suggestion.Category, suggestion.Description, suggestion.Impact)
            }
        }
    }
}

func main() {
    structuredOutputExamples()
}
```

## Provider Switching

Example of how to switch between different LLM providers dynamically.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

type LLMProvider interface {
    Client() *openai.Client
    GetDefaultModel() string
    ProviderName() string
}

// Extend providers to include name
type NamedGeminiInterface struct {
    *GeminiInterface
}

func (g *NamedGeminiInterface) ProviderName() string {
    return "Google Gemini"
}

type NamedOpenaiInterface struct {
    *OpenaiInterface
}

func (o *NamedOpenaiInterface) ProviderName() string {
    return "OpenAI"
}

func createProvider(providerName string) LLMProvider {
    switch providerName {
    case "gemini":
        return &NamedGeminiInterface{NewGemini(os.Getenv("GEMINI_API_KEY"))}
    case "openai":
        return &NamedOpenaiInterface{NewOpenai(os.Getenv("OPENAI_API_KEY"))}
    case "ollama":
        // Note: Ollama doesn't implement ProviderName in base, would need similar extension
        return NewOllama("http://localhost:11434/v1/", "", "llama2")
    default:
        return &NamedGeminiInterface{NewGemini(os.Getenv("GEMINI_API_KEY"))}
    }
}

func testWithMultipleProviders() {
    providers := []string{"gemini", "openai"}
    testQuestion := "Explain the concept of microservices architecture in 2-3 sentences."
    
    for _, providerName := range providers {
        fmt.Printf("\n=== Testing with %s ===\n", providerName)
        
        provider := createProvider(providerName)
        agent := NewAgent(
            context.Background(),
            provider.Client(),
            provider.GetDefaultModel(),
            "You are a helpful technical assistant.",
        )
        
        message := NewMessages()
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(testQuestion),
        ))
        
        if err != nil {
            log.Printf("Error with %s: %v", providerName, err)
            continue
        }
        
        fmt.Printf("Provider: %s\n", provider.ProviderName())
        fmt.Printf("Model: %s\n", provider.GetDefaultModel())
        fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
    }
}

func main() {
    testWithMultipleProviders()
}
```

## Error Handling Patterns

Comprehensive error handling examples for different scenarios.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

func robustWeatherTool(params map[string]string) string {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Weather tool panic recovered: %v", r)
        }
    }()
    
    location := params["location"]
    if location == "" {
        return `{"error": "Location parameter is required", "code": "MISSING_LOCATION"}`
    }
    
    // Simulate API failure
    if strings.ToLower(location) == "mars" {
        return `{"error": "Weather data not available for this location", "code": "LOCATION_NOT_FOUND"}`
    }
    
    return fmt.Sprintf(`{
        "temperature": "22°C",
        "condition": "sunny",
        "location": "%s",
        "success": true
    }`, location)
}

func errorHandlingExample() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a helpful assistant. Handle errors gracefully and provide useful feedback.",
    )
    
    // Add tool with error handling
    agent.AddTool(
        "get_weather",
        "Get weather information (may fail for some locations)",
        map[string]jsonschema.Definition{
            "location": {
                Type:        jsonschema.String,
                Description: "Location name",
            },
        },
        []string{"location"},
        robustWeatherTool,
    )
    
    message := NewMessages()
    
    // Test cases with different error scenarios
    testCases := []struct {
        query       string
        expectError bool
    }{
        {"What's the weather in London?", false},
        {"Tell me about weather on Mars", true},
        {"", false}, // Empty query
    }
    
    for i, testCase := range testCases {
        fmt.Printf("\n=== Test Case %d ===\n", i+1)
        fmt.Printf("Query: %s\n", testCase.query)
        
        // Add timeout context
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        
        // Create new agent with timeout context
        timeoutAgent := NewAgent(
            ctx,
            llm.Client(),
            llm.GetDefaultModel(),
            "You are a helpful assistant. Handle errors gracefully.",
        )
        
        // Re-add tool to timeout agent
        timeoutAgent.AddTool(
            "get_weather",
            "Get weather information",
            map[string]jsonschema.Definition{
                "location": {Type: jsonschema.String, Description: "Location name"},
            },
            []string{"location"},
            robustWeatherTool,
        )
        
        resp, err := timeoutAgent.Ask(message.MergeMessages(
            message.UserMessage(testCase.query),
        ))
        
        cancel() // Clean up context
        
        if err != nil {
            // Handle different types of errors
            switch {
            case strings.Contains(err.Error(), "context deadline exceeded"):
                fmt.Printf("Error: Request timed out\n")
            case strings.Contains(err.Error(), "authentication"):
                fmt.Printf("Error: Authentication failed - check API key\n")
            case strings.Contains(err.Error(), "rate limit"):
                fmt.Printf("Error: Rate limit exceeded - try again later\n")
            case strings.Contains(err.Error(), "tool call"):
                fmt.Printf("Error: Tool execution failed - %v\n", err)
            case strings.Contains(err.Error(), "maximum tool call depth"):
                fmt.Printf("Error: Too many nested tool calls\n")
            default:
                fmt.Printf("Error: %v\n", err)
            }
            
            if !testCase.expectError {
                fmt.Printf("Unexpected error for this test case\n")
            }
            continue
        }
        
        fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
        
        if testCase.expectError {
            fmt.Printf("Expected error but got success\n")
        }
    }
}

func main() {
    errorHandlingExample()
}
```

## Advanced Use Cases

### RAG (Retrieval-Augmented Generation) Pattern

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strings"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

// Mock knowledge base
var knowledgeBase = map[string]string{
    "go-channels": "Go channels are a typed conduit through which you can send and receive values with the channel operator <-. Channels provide a way for goroutines to communicate and synchronize.",
    "go-interfaces": "An interface type in Go is defined as a set of method signatures. A value of interface type can hold any value that implements those methods.",
    "go-goroutines": "A goroutine is a lightweight thread managed by the Go runtime. You create a new goroutine by putting 'go' in front of a function call.",
}

func searchKnowledgeTool(params map[string]string) string {
    query := strings.ToLower(params["query"])
    
    results := []map[string]string{}
    for key, content := range knowledgeBase {
        if strings.Contains(key, query) || strings.Contains(strings.ToLower(content), query) {
            results = append(results, map[string]string{
                "id":      key,
                "content": content,
            })
        }
    }
    
    response := map[string]interface{}{
        "query":   params["query"],
        "results": results,
        "count":   len(results),
    }
    
    jsonResult, _ := json.Marshal(response)
    return string(jsonResult)
}

func ragExample() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a Go programming expert. Use the search tool to find relevant information, then provide comprehensive answers based on the retrieved content.",
    )
    
    agent.AddTool(
        "search_knowledge",
        "Search the knowledge base for relevant information",
        map[string]jsonschema.Definition{
            "query": {
                Type:        jsonschema.String,
                Description: "Search query to find relevant information",
            },
        },
        []string{"query"},
        searchKnowledgeTool,
    )
    
    message := NewMessages()
    
    queries := []string{
        "How do channels work in Go?",
        "Explain Go interfaces",
        "What are the differences between goroutines and channels?",
    }
    
    for _, query := range queries {
        fmt.Printf("\n=== Query: %s ===\n", query)
        resp, err := agent.Ask(message.MergeMessages(
            message.UserMessage(query),
        ))
        
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        
        fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
    }
}

func main() {
    ragExample()
}
```

### Workflow Automation

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/sashabaranov/go-openai/jsonschema"
)

func workflowExample() {
    llm := NewGemini(os.Getenv("GEMINI_API_KEY"))
    agent := NewAgent(
        context.Background(),
        llm.Client(),
        llm.GetDefaultModel(),
        "You are a workflow automation assistant. Help users complete multi-step processes by breaking them down and executing each step.",
    )
    
    // Simulate various workflow steps
    agent.AddTool(
        "send_notification",
        "Send a notification message",
        map[string]jsonschema.Definition{
            "recipient": {Type: jsonschema.String, Description: "Notification recipient"},
            "message":   {Type: jsonschema.String, Description: "Notification message"},
            "channel":   {Type: jsonschema.String, Enum: []string{"email", "slack", "sms"}, Description: "Notification channel"},
        },
        []string{"recipient", "message"},
        func(params map[string]string) string {
            return fmt.Sprintf(`{"status": "sent", "recipient": "%s", "channel": "%s", "timestamp": "%s"}`,
                params["recipient"], params["channel"], time.Now().Format(time.RFC3339))
        },
    )
    
    agent.AddTool(
        "create_calendar_event",
        "Create a calendar event",
        map[string]jsonschema.Definition{
            "title":       {Type: jsonschema.String, Description: "Event title"},
            "date":        {Type: jsonschema.String, Description: "Event date (YYYY-MM-DD)"},
            "time":        {Type: jsonschema.String, Description: "Event time (HH:MM)"},
            "duration":    {Type: jsonschema.String, Description: "Event duration in minutes"},
            "attendees":   {Type: jsonschema.String, Description: "Comma-separated list of attendee emails"},
        },
        []string{"title", "date", "time"},
        func(params map[string]string) string {
            return fmt.Sprintf(`{"event_id": "evt_%d", "title": "%s", "date": "%s", "time": "%s", "status": "created"}`,
                time.Now().Unix(), params["title"], params["date"], params["time"])
        },
    )
    
    agent.AddTool(
        "update_project_status",
        "Update project status in management system",
        map[string]jsonschema.Definition{
            "project_id": {Type: jsonschema.String, Description: "Project identifier"},
            "status":     {Type: jsonschema.String, Enum: []string{"planning", "active", "on_hold", "completed"}, Description: "New project status"},
            "notes":      {Type: jsonschema.String, Description: "Status update notes"},
        },
        []string{"project_id", "status"},
        func(params map[string]string) string {
            return fmt.Sprintf(`{"project_id": "%s", "old_status": "active", "new_status": "%s", "updated_at": "%s"}`,
                params["project_id"], params["status"], time.Now().Format(time.RFC3339))
        },
    )
    
    message := NewMessages()
    
    // Complex workflow request
    workflowRequest := `
    I need to complete a project milestone. Please help me:
    1. Update project PRJ-123 status to completed
    2. Send a notification to the team via email about the completion
    3. Schedule a project retrospective meeting for next Friday at 2 PM
    4. Notify all stakeholders about the retrospective meeting
    `
    
    fmt.Printf("Workflow Request: %s\n", workflowRequest)
    fmt.Println("\n=== Executing Workflow ===")
    
    resp, err := agent.Ask(message.MergeMessages(
        message.UserMessage(workflowRequest),
    ))
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Workflow Result: %s\n", resp.Choices[0].Message.Content)
}

func main() {
    workflowExample()
}
```

## Running the Examples

To run any of these examples:

1. Set up your environment variables:
```bash
export GEMINI_API_KEY="your-gemini-key"
export OPENAI_API_KEY="your-openai-key"
```

2. Save any example to a Go file (e.g., `example.go`)

3. Initialize a Go module if needed:
```bash
go mod init example
go mod tidy
```

4. Run the example:
```bash
go run example.go
```

## Next Steps

- Explore the [Agent API documentation](agent.md)
- Learn about [Tool development](tool.md)
- Understand [Schema definitions](schema.md)
- Try different [LLM providers](llm.md)