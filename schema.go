package sapiens

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

func (as *Schema) ConvertSchema(s Schema) (*genai.Schema, error) {
	result := &genai.Schema{
		Type:        as.ParseType(s.Type),
		Format:      s.Format,
		Description: s.Description,
		Nullable:    s.Nullable,
		Enum:        s.Enum,
		Required:    s.Required,
	}

	if s.Items != nil {
		convertedItems, err := as.ConvertSchema(*s.Items) // Recursive call
		if err != nil {
			return nil, fmt.Errorf("error converting items: %w", err)
		}
		result.Items = convertedItems
	}

	if len(s.Properties) > 0 {
		result.Properties = make(map[string]*genai.Schema)
		for k, v := range s.Properties {
			convertedProperty, err := as.ConvertSchema(v) // Recursive call
			if err != nil {
				return nil, fmt.Errorf("error converting property %s: %w", k, err)
			}
			result.Properties[k] = convertedProperty
		}
	}

	return result, nil
}

func (as *Schema) ParseType(datatype string) genai.Type {
	switch datatype {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "object":
		return genai.TypeObject
	case "array":
		return genai.TypeArray
	default:
		return genai.TypeString
	}
}
