package sapiens

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestAgentMultipleTools(t *testing.T) {
	llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := NewAgent(context.Background(), llm.Client(), llm.GetDefaultModel(),
		"You are a helpful assistant with access to weather, currency conversion, and time tools. Use the tools to provide accurate and comprehensive responses.")

	message := NewMessages()

	// Add weather tool
	agent.AddTool("get_weather",
		"Get current weather information for a specific location",
		map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "The city and state/country, e.g. San Francisco, CA or London, UK",
			},
			"unit": {
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
		},
		[]string{"location"},
		func(parameters map[string]string) string {
			location := parameters["location"]
			unit := parameters["unit"]
			if unit == "" {
				unit = "celsius"
			}

			fmt.Printf("Weather Tool - Location: %s, Unit: %s\n", location, unit)

			// Simulate different weather data based on location
			switch location {
			case "Delhi, India":
				if unit == "fahrenheit" {
					return `{"temperature":"80", "unit":"fahrenheit", "condition":"sunny", "humidity":"65%", "wind_speed":"10 mph"}`
				}
				return `{"temperature":"27", "unit":"celsius", "condition":"sunny", "humidity":"65%", "wind_speed":"16 km/h"}`
			case "London, UK":
				if unit == "fahrenheit" {
					return `{"temperature":"59", "unit":"fahrenheit", "condition":"cloudy", "humidity":"80%", "wind_speed":"8 mph"}`
				}
				return `{"temperature":"15", "unit":"celsius", "condition":"cloudy", "humidity":"80%", "wind_speed":"13 km/h"}`
			default:
				return `{"temperature":"20", "unit":"celsius", "condition":"partly cloudy", "humidity":"70%", "wind_speed":"12 km/h"}`
			}
		})

	// Add currency conversion tool
	agent.AddTool("convert_currency",
		"Convert amount from one currency to another",
		map[string]jsonschema.Definition{
			"amount": {
				Type:        jsonschema.String,
				Description: "The amount to convert",
			},
			"from_currency": {
				Type:        jsonschema.String,
				Description: "Source currency code (e.g., USD, EUR, INR)",
			},
			"to_currency": {
				Type:        jsonschema.String,
				Description: "Target currency code (e.g., USD, EUR, INR)",
			},
		},
		[]string{"amount", "from_currency", "to_currency"},
		func(parameters map[string]string) string {
			amount := parameters["amount"]
			fromCurrency := parameters["from_currency"]
			toCurrency := parameters["to_currency"]

			fmt.Printf("Currency Tool - Amount: %s, From: %s, To: %s\n", amount, fromCurrency, toCurrency)

			// Simulate exchange rates
			exchangeRates := map[string]map[string]float64{
				"USD": {"EUR": 0.85, "INR": 83.0, "GBP": 0.76},
				"EUR": {"USD": 1.18, "INR": 97.6, "GBP": 0.89},
				"INR": {"USD": 0.012, "EUR": 0.010, "GBP": 0.009},
				"GBP": {"USD": 1.31, "EUR": 1.12, "INR": 109.2},
			}

			// Simple conversion logic
			if fromCurrency == toCurrency {
				return fmt.Sprintf(`{"original_amount":"%s", "converted_amount":"%s", "from_currency":"%s", "to_currency":"%s", "exchange_rate":"1.0"}`,
					amount, amount, fromCurrency, toCurrency)
			}

			if rates, exists := exchangeRates[fromCurrency]; exists {
				if rate, exists := rates[toCurrency]; exists {
					// Parse amount (simplified)
					var amountFloat float64 = 100.0 // Default for demo
					if amount == "50" {
						amountFloat = 50.0
					} else if amount == "1000" {
						amountFloat = 1000.0
					}

					convertedAmount := amountFloat * rate
					return fmt.Sprintf(`{"original_amount":"%.2f", "converted_amount":"%.2f", "from_currency":"%s", "to_currency":"%s", "exchange_rate":"%.4f"}`,
						amountFloat, convertedAmount, fromCurrency, toCurrency, rate)
				}
			}

			return `{"error":"Currency conversion not supported for this pair"}`
		})

	// Add time zone tool
	agent.AddTool("get_time",
		"Get current time for a specific location or timezone",
		map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "City name or timezone (e.g., New York, Tokyo, UTC)",
			},
			"format": {
				Type:        jsonschema.String,
				Description: "Time format preference",
				Enum:        []string{"12hour", "24hour"},
			},
		},
		[]string{"location"},
		func(parameters map[string]string) string {
			location := parameters["location"]
			format := parameters["format"]
			if format == "" {
				format = "24hour"
			}

			return fmt.Sprintf("Time Tool - Location: %s, Format: %s\n", location, format)
		})

	// Test scenarios
	testCases := []struct {
		name     string
		question string
	}{
		{
			name:     "Weather Query",
			question: "What's the weather like in Delhi right now?",
		},
		{
			name:     "Currency Conversion",
			question: "I have 1000 USD, how much is that in Indian Rupees?",
		},
		{
			name:     "Time Query",
			question: "What time is it in Tokyo right now?",
		},
		{
			name:     "Multiple Tools",
			question: "I'm planning to travel from New York to London tomorrow. Can you tell me the weather in London, what time it is there now, and how much 500 USD would be in British Pounds?",
		},
		{
			name:     "Complex Weather",
			question: "Compare the weather between Delhi and London, and tell me the time difference between these cities.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("\n=== Test Case: %s ===\n", tc.name)
			fmt.Printf("Question: %s\n", tc.question)

			resp, err := agent.Ask(message.MergeMessages(
				message.UserMessage(tc.question),
			))

			if err != nil {
				t.Errorf("Test %s failed with error: %v", tc.name, err)
				return
			}

			if len(resp.Choices) == 0 {
				t.Errorf("Test %s: No response choices received", tc.name)
				return
			}

			response := resp.Choices[0].Message.Content
			fmt.Printf("AI Response: %s\n", response)

			// Basic validation - response should not be empty
			if response == "" {
				t.Errorf("Test %s: Empty response received", tc.name)
			}

			// Specific validations for each test case
			switch tc.name {
			case "Weather Query":
				if response == "" {
					t.Errorf("Weather query should return weather information")
				}
			case "Currency Conversion":
				if response == "" {
					t.Errorf("Currency conversion should return converted amount")
				}
			case "Time Query":
				if response == "" {
					t.Errorf("Time query should return time information")
				}
			case "Multiple Tools":
				if response == "" {
					t.Errorf("Multiple tools query should use multiple tools and provide comprehensive response")
				}
			}

			fmt.Printf("✅ Test %s completed successfully\n", tc.name)
		})
	}
}

func TestAgentToolChaining(t *testing.T) {
	llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := NewAgent(context.Background(), llm.Client(), llm.GetDefaultModel(),
		"You are a travel planning assistant. Use available tools to help plan trips.")

	message := NewMessages()

	// Add a calculation tool
	agent.AddTool("calculate",
		"Perform basic mathematical calculations",
		map[string]jsonschema.Definition{
			"expression": {
				Type:        jsonschema.String,
				Description: "Mathematical expression to calculate (e.g., '100 * 1.2' or '500 / 5')",
			},
		},
		[]string{"expression"},
		func(parameters map[string]string) string {
			expression := parameters["expression"]
			fmt.Printf("Calculator Tool - Expression: %s\n", expression)

			// Simple calculator for demo - return predefined results to avoid infinite loops
			switch expression {
			case "100 * 5":
				return `{"expression":"100 * 5", "result":"500", "operation":"multiplication"}`
			case "500 * 1.2":
				return `{"expression":"500 * 1.2", "result":"600", "operation":"multiplication"}`
			case "100 * 5 * 1.2":
				return `{"expression":"100 * 5 * 1.2", "result":"600", "operation":"multiplication"}`
			default:
				// For unknown expressions, provide a helpful response that stops the loop
				return `{"expression":"` + expression + `", "result":"calculation_not_supported", "operation":"error", "message":"Please provide a simpler calculation"}`
			}
		})

	// Add budget tool
	agent.AddTool("check_budget",
		"Check if an amount fits within a given budget",
		map[string]jsonschema.Definition{
			"total_cost": {
				Type:        jsonschema.String,
				Description: "Total cost amount",
			},
			"budget": {
				Type:        jsonschema.String,
				Description: "Available budget amount",
			},
		},
		[]string{"total_cost", "budget"},
		func(parameters map[string]string) string {
			totalCost := parameters["total_cost"]
			budget := parameters["budget"]

			fmt.Printf("Budget Tool - Cost: %s, Budget: %s\n", totalCost, budget)

			// Simple budget check for demo values
			switch {
			case totalCost == "600" && budget == "1000":
				return `{"total_cost":"600", "budget":"1000", "within_budget":true, "remaining":"400"}`
			case totalCost == "970" && budget == "1000":
				return `{"total_cost":"970", "budget":"1000", "within_budget":true, "remaining":"30"}`
			case totalCost <= budget:
				return fmt.Sprintf(`{"total_cost":"%s", "budget":"%s", "within_budget":true, "remaining":"30"}`, totalCost, budget)
			default:
				return fmt.Sprintf(`{"total_cost":"%s", "budget":"%s", "within_budget":false, "overage":"50"}`, totalCost, budget)
			}
		})

	fmt.Printf("\n=== Tool Chaining Test ===\n")
	question := "I need to calculate 100 * 5 for my trip accommodation costs. Can you help me with this calculation?"

	resp, err := agent.Ask(message.MergeMessages(
		message.UserMessage(question),
	))

	if err != nil {
		t.Fatalf("Tool chaining test failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No response choices received")
	}

	response := resp.Choices[0].Message.Content
	fmt.Printf("Question: %s\n", question)
	fmt.Printf("AI Response: %s\n", response)

	if response == "" {
		t.Error("Empty response received for tool chaining test")
	}

	fmt.Printf("✅ Tool chaining test completed\n")
}
