package sapiens

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestAgentAsk(t *testing.T) {

	llm := NewGemini(os.Getenv("GEMINI_API_KEY"))

	agent := NewAgent(context.Background(), llm.Client(), llm.GetDefaultModel(),
		"you are a weather reporter")

	message := NewMessages()

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

	resp, err := agent.Ask(
		message.MergeMessages(
			message.UserMessage("what is the weather in delhi and convert 1 inr to usd"),
		),
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)

}
