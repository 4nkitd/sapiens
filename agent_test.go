package sapiens

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAgent(t *testing.T) {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()
	response, err := agent.Run(ctx, "What is the capital of France ?")
	if err != nil {
		t.Fatalf("Agent.Run failed: %v", err)
	}

	t.Logf("Response: %s", response.Content)

	if response.Content == "" {
		t.Fatal("Expected non-empty response but got empty string")
	}
}

func TestAgentWithToolCalling(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()

	// Test with tools
	weatherTool := Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &Schema{
			Type: "object",
			Properties: map[string]Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit (celsius or fahrenheit)",
					Enum:        []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"location"},
		},
	}

	agent.AddTools(weatherTool)

	toolResponse, err := agent.Run(ctx, "What's the weather like in Paris?")
	if err != nil {
		t.Fatalf("Agent.Run with tools failed: %v", err)
	}

	t.Logf("Tool calls: %v", toolResponse.ToolCalls)
	t.Logf("Content: %s", toolResponse.Content)

}

func TestAgentWithStructuredResponse(t *testing.T) {
	apiKey :=
		os.Getenv("GEMINI_API_KEY")

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate and concise answers.", "1.0")

	// Test basic conversation
	ctx := context.Background()

	// Test with structured output
	agent = NewAgent("StructuredAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate responses in structured format.", "1.0")

	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"failure_reason": {
				Type:        "string",
				Description: "Reason for failure",
			},
			"answer": {
				Type:        "string",
				Description: "The answer to the user's question",
			},
			"confidence": {
				Type:        "number",
				Description: "Confidence score from 0 to 1",
			},
		},
		Required: []string{"answer", "confidence"},
	}

	agent.SetStructuredResponseSchema(schema)

	structuredResponse, err := agent.Run(ctx, "what is the population india?")
	fmt.Println(err)
	if err != nil {
		t.Fatalf("Agent.Run with structured output failed: %v", err)
	}

	t.Logf("Structured: %v", structuredResponse.Structured)
	t.Logf("Content: %s", structuredResponse.Content)

	if structuredResponse.Structured == nil {
		t.Fatal("Expected structured data but got nil")
	}

}

func TestAgentWithToolImplementation(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("TestAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant.", "1.0")

	// Test with tools and implementation
	weatherTool := Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &Schema{
			Type: "object",
			Properties: map[string]Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit (celsius or fahrenheit)",
					Enum:        []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"location"},
		},
	}

	agent.AddTools(weatherTool)

	// Register tool implementation
	agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, ok := params["location"].(string)
		if !ok {
			return nil, fmt.Errorf("location must be a string")
		}

		unit := "celsius"
		if unitParam, ok := params["unit"].(string); ok {
			unit = unitParam
		}

		// Mock weather data
		weatherInfo := map[string]interface{}{
			"location":    location,
			"temperature": 22,
			"unit":        unit,
			"condition":   "Rainy",
		}

		fmt.Printf("Weather info: %v\n", weatherInfo)

		return weatherInfo, nil
	})

	ctx := context.Background()
	response, err := agent.Run(ctx, "What's the weather like in Paris?")
	if err != nil {
		t.Fatalf("Agent.Run with tool implementation failed: %v", err)
	}

	t.Logf("Final response: %s", response.Content)

	// Test a follow-up question to see if the agent can use the context
	followUpResponse, err := agent.Run(ctx, "And what should I wear in that weather?")
	if err != nil {
		t.Fatalf("Follow-up question failed: %v", err)
	}

	t.Logf("Follow-up response: %s", followUpResponse.Content)
}

func TestAgentMemory(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("MemoryAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant who remembers previous parts of the conversation.", "1.0")

	ctx := context.Background()

	// Initial message with information to remember
	initialResponse, err := agent.Run(ctx, "My name is Alice and I have a dog named Max.")
	if err != nil {
		t.Fatalf("Initial message failed: %v", err)
	}
	t.Logf("Initial response: %s", initialResponse.Content)

	// First follow-up question that requires memory
	followUpResponse1, err := agent.Run(ctx, "What's my name?")
	if err != nil {
		t.Fatalf("First follow-up question failed: %v", err)
	}
	t.Logf("First follow-up response: %s", followUpResponse1.Content)

	// Check if the response contains the name
	if !containsSubstring(followUpResponse1.Content, "Alice") {
		t.Errorf("Expected response to contain 'Alice', but got: %s", followUpResponse1.Content)
	}

	// Second follow-up to test deeper memory
	followUpResponse2, err := agent.Run(ctx, "What is my dog's name?")
	if err != nil {
		t.Fatalf("Second follow-up question failed: %v", err)
	}
	t.Logf("Second follow-up response: %s", followUpResponse2.Content)

	// Check if the response contains the dog's name
	if !containsSubstring(followUpResponse2.Content, "Max") {
		t.Errorf("Expected response to contain 'Max', but got: %s", followUpResponse2.Content)
	}

	// Test memory persistence with additional context
	furtherResponse, err := agent.Run(ctx, "I also have a cat named Luna. Can you remember both my pets?")
	if err != nil {
		t.Fatalf("Further context message failed: %v", err)
	}
	t.Logf("Response with additional context: %s", furtherResponse.Content)

	// Final memory check
	finalResponse, err := agent.Run(ctx, "What are the names of my pets?")
	if err != nil {
		t.Fatalf("Final memory check failed: %v", err)
	}
	t.Logf("Final memory check response: %s", finalResponse.Content)

	// Check if both pet names are mentioned
	if !containsSubstring(finalResponse.Content, "Max") || !containsSubstring(finalResponse.Content, "Luna") {
		t.Errorf("Expected response to contain both 'Max' and 'Luna', but got: %s", finalResponse.Content)
	}
}

// Helper function to check if a string contains a substring (case insensitive)
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestAgentWithInitialContext(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("ContextAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that answers questions based on the provided context.", "1.0")

	ctx := context.Background()

	// Prepare initial context about a fictional company
	initialContext := `
		Company: TechNova Solutions
		Founded: 2015
		CEO: Sarah Johnson
		Headquarters: San Francisco, CA
		Employees: 450
		Main Products: 
		- CloudSync (cloud storage solution)
		- DataGuard (cybersecurity platform)
		- SmartAnalytics (business intelligence tool)
		Recent News: TechNova Solutions just raised $50 million in Series C funding led by Acme Ventures.
	`

	// Set the initial context for the agent as a string
	agent.SetStringContext(initialContext)

	// Ask questions that should be answered based on the context
	questions := []string{
		"Who is the CEO of the company?",
		"What products does the company offer?",
		"Where is the company headquartered?",
		"What was the recent funding news?",
	}

	expectedAnswers := []string{
		"Sarah Johnson",
		"CloudSync", "DataGuard", "SmartAnalytics",
		"San Francisco",
		"$50 million", "Series C", "Acme Ventures",
	}

	for i, question := range questions {
		response, err := agent.Run(ctx, question)
		if err != nil {
			t.Fatalf("Question #%d failed: %v", i+1, err)
		}

		t.Logf("Question: %s\nResponse: %s", question, response.Content)

		// Check if the response contains expected information
		for _, expectedText := range expectedAnswers {
			if strings.Contains(question, strings.Split(expectedText, " ")[0]) && !containsSubstring(response.Content, expectedText) {
				t.Errorf("Question #%d: Expected response to contain '%s', but got: %s", i+1, expectedText, response.Content)
			}
		}
	}

	// Test asking about something not in the context
	unrelatedResponse, err := agent.Run(ctx, "What is the capital of France?")
	if err != nil {
		t.Fatalf("Unrelated question failed: %v", err)
	}

	t.Logf("Unrelated question response: %s", unrelatedResponse.Content)

	// Test context update with a string
	updateContext := `
		Updated Information:
		CEO: Michael Chen (promoted from CTO in 2023)
		Employees: Now 520 after recent expansion
		New Product: EdgeCompute (edge computing platform) launched last month
	`

	agent.UpdateStringContext(updateContext)

	// Test with updated context
	updatedResponse, err := agent.Run(ctx, "Who is the current CEO and when did they take the position?")
	if err != nil {
		t.Fatalf("Question after context update failed: %v", err)
	}

	t.Logf("Updated context response: %s", updatedResponse.Content)

	if !containsSubstring(updatedResponse.Content, "Michael Chen") || !containsSubstring(updatedResponse.Content, "2023") {
		t.Errorf("Expected response to mention 'Michael Chen' and '2023', but got: %s", updatedResponse.Content)
	}
}

func TestAgentWithDynamicPrompt(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("DynamicPromptAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")

	// Define a prompt template with placeholders for dynamic content
	promptTemplate := `
	You are a knowledgeable assistant specialized in {{field}} topics.
	
	When discussing {{topic}}, please keep these key points in mind:
	{{#points}}
	- {{.}}
	{{/points}}
	
	{{#if examples}}
	Here are some examples to consider:
	{{#examples}}
	* {{name}}: {{description}}
	{{/examples}}
	{{/if}}
	
	Today's focus is on {{focus}}.
	`

	// Create dynamic card data
	cardData := map[string]interface{}{
		"field": "technology",
		"topic": "artificial intelligence",
		"points": []string{
			"Ethical considerations are important",
			"Technology should be accessible to everyone",
			"Privacy must be respected",
		},
		"examples": []map[string]string{
			{"name": "Machine Learning", "description": "Training systems on data to make predictions"},
			{"name": "Computer Vision", "description": "Enabling machines to interpret visual information"},
			{"name": "Natural Language Processing", "description": "Helping computers understand human language"},
		},
		"focus": "developing responsible AI systems",
	}

	// Apply the dynamic card data to the prompt template
	dynamicPrompt, err := ApplyTemplate(promptTemplate, cardData)
	if err != nil {
		t.Fatalf("Failed to apply template: %v", err)
	}

	// Add the dynamic prompt as the system prompt
	agent.AddSystemPrompt(dynamicPrompt, "1.0")

	// Test the agent with a relevant question
	ctx := context.Background()
	response, err := agent.Run(ctx, "What are the most important considerations when developing AI systems?")
	if err != nil {
		t.Fatalf("Agent.Run failed: %v", err)
	}

	t.Logf("Response: %s", response.Content)

	// Verify the response contains relevant information from the dynamic prompt
	expectedTerms := []string{"ethical", "accessibility", "privacy", "responsible"}

	foundTerms := 0
	for _, term := range expectedTerms {
		if containsSubstring(response.Content, term) {
			foundTerms++
		}
	}

	// Assert that at least 2 of the expected terms are mentioned
	if foundTerms < 2 {
		t.Errorf("Expected response to mention at least 2 terms from %v, but found only %d terms",
			expectedTerms, foundTerms)
	}

	// Test modifying the card data and updating the prompt
	updatedCardData := map[string]interface{}{
		"field": "technology",
		"topic": "artificial intelligence",
		"points": []string{
			"AI should be explainable and transparent",
			"Models should be regularly evaluated for bias",
			"Human oversight is necessary for critical decisions",
		},
		"examples": []map[string]string{
			{"name": "Healthcare AI", "description": "Using AI to improve patient outcomes"},
			{"name": "Financial AI", "description": "AI systems for fraud detection and risk assessment"},
		},
		"focus": "ethical AI implementation strategies",
	}

	// Apply the updated card data
	updatedPrompt, err := ApplyTemplate(promptTemplate, updatedCardData)
	if err != nil {
		t.Fatalf("Failed to apply updated template: %v", err)
	}

	// Update the system prompt with the new dynamic content
	agent.AddSystemPrompt(updatedPrompt, "1.1")

	// Test with a new question
	updatedResponse, err := agent.Run(ctx, "How can we ensure AI systems make fair decisions?")
	if err != nil {
		t.Fatalf("Agent.Run with updated prompt failed: %v", err)
	}

	t.Logf("Updated Response: %s", updatedResponse.Content)

	// Verify the response contains relevant information from the updated prompt
	updatedExpectedTerms := []string{"transparent", "bias", "explainable", "oversight", "ethical"}

	updatedFoundTerms := 0
	for _, term := range updatedExpectedTerms {
		if containsSubstring(updatedResponse.Content, term) {
			updatedFoundTerms++
		}
	}

	// Assert that at least 2 of the expected terms are mentioned
	if updatedFoundTerms < 2 {
		t.Errorf("Expected updated response to mention at least 2 terms from %v, but found only %d terms",
			updatedExpectedTerms, updatedFoundTerms)
	}
}

func TestAgentWithPromptManager(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("PromptManagerAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")

	// Define a prompt template for a customer service agent
	customerServiceTemplate := PromptTemplate{
		Name: "customer_service",
		Template: `
		You are a {{tone}} customer service representative for {{company_name}}, a company that specializes in {{industry}}.

		When helping customers:
		{{#guidelines}}
		- {{.}}
		{{/guidelines}}

		Company policies you must follow:
		{{#policies}}
		* {{.}}
		{{/policies}}

		Today's focus is on {{daily_focus}}.
		`,
		Description: "Template for customer service representative prompts",
		Version:     "1.0",
	}

	// Add the template to the agent's prompt manager
	err = agent.AddPromptTemplate(customerServiceTemplate)
	if err != nil {
		t.Fatalf("Failed to add prompt template: %v", err)
	}

	// Create a card with data for this prompt
	cardData := map[string]interface{}{
		"tone":         "friendly and helpful",
		"company_name": "TechSupport Pro",
		"industry":     "technical support for software products",
		"guidelines": []string{
			"Always greet the customer by name when provided",
			"Use simple, non-technical language unless the customer demonstrates technical knowledge",
			"Offer step-by-step guidance for resolving issues",
			"Verify the solution worked before ending the conversation",
		},
		"policies": []string{
			"Do not share customer data with third parties",
			"Refunds can be processed within 30 days of purchase",
			"Premium support is available for subscribers only",
			"Service hours are 9AM to 6PM Monday through Friday",
		},
		"daily_focus": "reducing resolution time while maintaining customer satisfaction",
	}

	// Create a card using the template and data
	card := NewCard("customer_service", cardData)

	// Add the card as a system prompt
	err = agent.AddDynamicPromptWithCard(card, "1.0")
	if err != nil {
		t.Fatalf("Failed to add dynamic prompt with card: %v", err)
	}

	// Test the agent with a customer service question
	ctx := context.Background()
	response, err := agent.Run(ctx, "Hi, I'm having trouble installing your software. It gets stuck at 75% and then shows an error about missing dependencies.")
	if err != nil {
		t.Fatalf("Agent.Run failed: %v", err)
	}

	t.Logf("Response: %s", response.Content)

	// Verify the response is helpful and follows the guidelines
	expectedTerms := []string{"install", "dependencies", "step", "resolution", "error"}

	foundTerms := 0
	for _, term := range expectedTerms {
		if containsSubstring(response.Content, term) {
			foundTerms++
		}
	}

	if foundTerms < 3 {
		t.Errorf("Expected response to mention at least 3 terms from %v, but found only %d terms",
			expectedTerms, foundTerms)
	}

	// Test updating the card with new data
	updatedCardData := map[string]interface{}{
		"tone":         "professional and concise",
		"company_name": "TechSupport Pro",
		"industry":     "technical support for software products",
		"guidelines": []string{
			"Focus on efficiency and accurate solutions",
			"Provide troubleshooting steps in numbered format",
			"Include links to relevant documentation when possible",
			"Recommend preventative measures to avoid future issues",
		},
		"policies": []string{
			"Support ticket response times must be within 2 hours",
			"Critical issues must be escalated to the technical team",
			"All interactions must be documented in the customer record",
		},
		"daily_focus": "addressing installation and dependency-related issues",
	}

	// Create an updated card
	updatedCard := NewCard("customer_service", updatedCardData)

	// Update the system prompt with the new card
	err = agent.AddDynamicPromptWithCard(updatedCard, "1.1")
	if err != nil {
		t.Fatalf("Failed to update dynamic prompt with card: %v", err)
	}

	// Test with a similar question to see if the response style changes
	updatedResponse, err := agent.Run(ctx, "I'm still having issues with installing your software. Can you help?")
	if err != nil {
		t.Fatalf("Agent.Run with updated card failed: %v", err)
	}

	t.Logf("Updated Response: %s", updatedResponse.Content)

	// Verify the response follows the updated guidelines
	updatedExpectedTerms := []string{"1.", "2.", "documentation", "prevent", "installation"}

	updatedFoundTerms := 0
	for _, term := range updatedExpectedTerms {
		if containsSubstring(updatedResponse.Content, term) {
			updatedFoundTerms++
		}
	}

	if updatedFoundTerms < 2 {
		t.Errorf("Expected updated response to mention at least 2 terms from %v, but found only %d terms",
			updatedExpectedTerms, updatedFoundTerms)
	}
}

// Helper function to augment the prompt with memory
func augmentPromptWithMemory(agent *Agent, llm LLMInterface, prompt string) string {
	// Generate embedding for the prompt
	embedding, err := llm.GenerateEmbedding(context.Background(), "gemini-embedding-exp-03-07", prompt, SEMANTIC_SIMILARITY)
	if err != nil {
		fmt.Printf("Failed to generate embedding for prompt: %v\n", err)
		return prompt // Return original prompt on error
	}

	// Search memory for relevant information
	results := agent.Memory.Search(embedding.Vector)

	// Append memory results to the prompt
	for _, result := range results {
		key, ok := result.Key.(string) // Type assertion
		if !ok {
			fmt.Printf("Invalid key type in memory: %T\n", result.Key)
			continue // Skip this result
		}
		value := agent.Memory.Get(key) // Retrieve the value
		if value != nil {
			prompt += fmt.Sprintf("\nMemory: key=%s value=%v", key, value) // Add key and value
		} else {
			prompt += "\nMemory: " + result.Text // Fallback to text if value is nil
		}
	}

	return prompt
}

func TestAgentMemoryWithGemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("GeminiMemoryAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant who remembers previous parts of the conversation.", "1.0")

	// Create a new memory instance
	memory := NewMemory("simple", map[string]interface{}{})
	agent.Memory = &memory

	ctx := context.Background()

	// Information to remember
	name := "Bob"
	location := "New York"
	profession := "software engineer"

	// Generate embeddings for the information
	embeddingName, err := llmImpl.GenerateEmbedding(ctx, "gemini-embedding-exp-03-07", fmt.Sprintf("My name is %s", name), SEMANTIC_SIMILARITY)
	if err != nil {
		t.Fatalf("Failed to generate embedding for name: %v", err)
	}
	agent.Memory.Add("name", name, embeddingName)

	embeddingLocation, err := llmImpl.GenerateEmbedding(ctx, "gemini-embedding-exp-03-07", fmt.Sprintf("I live in %s", location), SEMANTIC_SIMILARITY)
	if err != nil {
		t.Fatalf("Failed to generate embedding for location: %v", err)
	}
	agent.Memory.Add("location", location, embeddingLocation)

	embeddingProfession, err := llmImpl.GenerateEmbedding(ctx, "gemini-embedding-exp-03-07", fmt.Sprintf("I work as a %s", profession), SEMANTIC_SIMILARITY)
	if err != nil {
		t.Fatalf("Failed to generate embedding for profession: %v", err)
	}
	agent.Memory.Add("profession", profession, embeddingProfession)

	// First follow-up question that requires memory
	prompt1 := "Where do I live?"
	augmentedPrompt1 := augmentPromptWithMemory(agent, llmImpl, prompt1)
	followUpResponse1, err := agent.Run(ctx, augmentedPrompt1)
	if err != nil {
		t.Fatalf("First follow-up question failed: %v", err)
	}
	t.Logf("First follow-up response: %s", followUpResponse1.Content)

	// Check if the response contains the location
	if !containsSubstring(followUpResponse1.Content, "New York") {
		t.Errorf("Expected response to contain 'New York', but got: %s", followUpResponse1.Content)
	}

	// Second follow-up to test deeper memory
	prompt2 := "What is my name?"
	augmentedPrompt2 := augmentPromptWithMemory(agent, llmImpl, prompt2)
	followUpResponse2, err := agent.Run(ctx, augmentedPrompt2)
	if err != nil {
		t.Fatalf("Second follow-up question failed: %v", err)
	}
	t.Logf("Second follow-up response: %s", followUpResponse2.Content)

	// Check if the response contains the name
	if !containsSubstring(followUpResponse2.Content, "Bob") {
		t.Errorf("Expected response to contain 'Bob', but got: %s", followUpResponse2.Content)
	}

	// Final memory check
	prompt3 := "What is my name, where do I live, and what is my profession?"
	augmentedPrompt3 := augmentPromptWithMemory(agent, llmImpl, prompt3)
	finalResponse, err := agent.Run(ctx, augmentedPrompt3)
	if err != nil {
		t.Fatalf("Final memory check failed: %v", err)
	}
	t.Logf("Final memory check response: %s", finalResponse.Content)

	// Check if all information is mentioned
	if !containsSubstring(finalResponse.Content, "Bob") || !containsSubstring(finalResponse.Content, "New York") || !containsSubstring(finalResponse.Content, "software engineer") {
		t.Errorf("Expected response to contain 'Bob', 'New York', and 'software engineer', but got: %s", finalResponse.Content)
	}
}

func TestAgentWithStructuredAndToolCalls(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	// Create a new LLM implementation
	llmImpl := NewGoogleGenAI(apiKey, "gemini-2.0-flash")
	err := llmImpl.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize GoogleGenAI: %v", err)
	}

	// Create a new agent
	agent := NewAgent("CombinedAgent", llmImpl, apiKey, "gemini-2.0-flash", "google")
	agent.AddSystemPrompt("You are a helpful AI assistant that provides accurate responses in structured format and can use tools.", "1.0")

	// Define structured response schema
	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"location": {
				Type:        "string",
				Description: "The location the user asked about",
			},
			"analysis": {
				Type:        "string",
				Description: "Analysis of the weather conditions",
			},
			"recommendation": {
				Type:        "string",
				Description: "Recommendation based on the weather",
			},
			"confidence": {
				Type:        "number",
				Description: "Confidence score from 0 to 1",
			},
		},
		Required: []string{"location", "analysis", "recommendation"},
	}

	// Set structured response schema
	agent.SetStructuredResponseSchema(schema)

	// Add weather tool
	weatherTool := Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		InputSchema: &Schema{
			Type: "object",
			Properties: map[string]Schema{
				"location": {
					Type:        "string",
					Description: "The city and state/country",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit (celsius or fahrenheit)",
					Enum:        []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"location"},
		},
	}

	agent.AddTools(weatherTool)

	// Register tool implementation
	agent.RegisterToolImplementation("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, ok := params["location"].(string)
		if !ok {
			return nil, fmt.Errorf("location must be a string")
		}

		unit := "celsius"
		if unitParam, ok := params["unit"].(string); ok {
			unit = unitParam
		}

		// Mock weather data
		weatherInfo := map[string]interface{}{
			"location":    location,
			"temperature": 22,
			"unit":        unit,
			"condition":   "Partly cloudy",
			"humidity":    65,
			"wind_speed":  12,
		}

		return weatherInfo, nil
	})

	ctx := context.Background()
	response, err := agent.Run(ctx, "What's the weather like in Tokyo and what should I wear?")
	if err != nil {
		t.Fatalf("Agent.Run with structured output and tools failed: %v", err)
	}

	// Log the response for debugging
	t.Logf("Tool calls: %v", response.ToolCalls)
	t.Logf("Structured: %v", response.Structured)
	t.Logf("Content: %s", response.Content)

	// Check if we have tool calls
	if len(response.ToolCalls) == 0 {
		t.Error("Expected tool calls but got none")
	}

	// Check if we have a weather tool call
	foundWeatherTool := false
	for _, call := range response.ToolCalls {
		if call.Name == "get_weather" {
			foundWeatherTool = true
			break
		}
	}
	if !foundWeatherTool {
		t.Error("Expected 'get_weather' tool call but didn't find it")
	}

	// Check if we have structured data
	if response.Structured == nil {
		t.Error("Expected structured data but got nil")
	} else {
		// Check for required fields in the structured response
		structMap, ok := response.Structured.(map[string]interface{})
		if !ok {
			t.Error("Structured response is not a map")
		} else {
			requiredFields := []string{"location", "analysis", "recommendation"}
			for _, field := range requiredFields {
				if _, exists := structMap[field]; !exists {
					t.Errorf("Required field '%s' missing from structured response", field)
				}
			}

			// Verify location matches what was asked
			if loc, exists := structMap["location"]; exists {
				locStr, ok := loc.(string)
				if !ok {
					t.Error("Location field is not a string")
				} else if !strings.Contains(strings.ToLower(locStr), "tokyo") {
					t.Errorf("Expected location to contain 'Tokyo', got '%s'", locStr)
				}
			}
		}
	}

	// Check if the content contains relevant information
	expectedTerms := []string{"Tokyo", "weather", "wear", "temperature"}
	foundTerms := 0
	for _, term := range expectedTerms {
		if containsSubstring(response.Content, term) {
			foundTerms++
		}
	}
	if foundTerms < 2 {
		t.Errorf("Expected response to mention at least 2 terms from %v, but found only %d terms",
			expectedTerms, foundTerms)
	}
}
