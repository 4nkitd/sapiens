package sapiens

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
)

// NewSystemPrompt creates a new system prompt
func NewSystemPrompt(prompt string, version string) Prompt {
	return Prompt{
		Prompt:  prompt,
		Version: version,
	}
}

// SetEnhanced adds an enhanced version of the prompt
func (s *Prompt) SetEnhanced(enhanced Prompt) {
	s.Enhanced = append(s.Enhanced, enhanced)
}

// GetLatestEnhanced returns the most recent enhanced version of the prompt
func (s *Prompt) GetLatestEnhanced() Prompt {
	if len(s.Enhanced) == 0 {
		return Prompt{}
	}
	return s.Enhanced[len(s.Enhanced)-1]
}

// Parse applies parameters to a prompt template
func (s *Prompt) Parse(raw string, parameters map[string]string) (output Prompt, err error) {
	prompt := raw
	if len(s.Enhanced) > 0 {
		prompt = s.GetLatestEnhanced().Prompt
	}

	// Create a new template from the raw string
	prompTmpl, err := template.New("prompt").Parse(prompt)
	if err != nil {
		return output, err
	}

	// Execute the template with the parameters
	var buf bytes.Buffer
	if err := prompTmpl.Execute(&buf, parameters); err != nil {
		return output, err
	}

	// Convert the rendered prompt to SystemPrompt
	output.Prompt = buf.String()
	return output, nil
}

// PromptTemplate represents a reusable prompt template with placeholders
type PromptTemplate struct {
	Name        string                 // Unique identifier for the template
	Template    string                 // The template string with placeholders
	Description string                 // Description of what the template is for
	Version     string                 // Version number of the template
	Metadata    map[string]interface{} // Additional metadata
}

// PromptManager handles storage and retrieval of prompt templates
type PromptManager struct {
	templates map[string]PromptTemplate
}

// NewPromptManager creates a new prompt manager instance
func NewPromptManager() *PromptManager {
	return &PromptManager{
		templates: make(map[string]PromptTemplate),
	}
}

// AddTemplate adds a new prompt template to the manager
func (pm *PromptManager) AddTemplate(template PromptTemplate) error {
	if _, exists := pm.templates[template.Name]; exists {
		return fmt.Errorf("template with name '%s' already exists", template.Name)
	}

	pm.templates[template.Name] = template
	return nil
}

// UpdateTemplate updates an existing template
func (pm *PromptManager) UpdateTemplate(template PromptTemplate) error {
	if _, exists := pm.templates[template.Name]; !exists {
		return fmt.Errorf("template with name '%s' does not exist", template.Name)
	}

	pm.templates[template.Name] = template
	return nil
}

// GetTemplate retrieves a template by name
func (pm *PromptManager) GetTemplate(name string) (PromptTemplate, error) {
	template, exists := pm.templates[name]
	if !exists {
		return PromptTemplate{}, fmt.Errorf("template with name '%s' not found", name)
	}

	return template, nil
}

// ListTemplates returns all available templates
func (pm *PromptManager) ListTemplates() []PromptTemplate {
	templates := make([]PromptTemplate, 0, len(pm.templates))
	for _, template := range pm.templates {
		templates = append(templates, template)
	}
	return templates
}

// RenderTemplate applies data to a template and returns the rendered prompt
func (pm *PromptManager) RenderTemplate(templateName string, data map[string]interface{}) (string, error) {
	template, err := pm.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	renderedPrompt, err := ApplyTemplate(template.Template, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template '%s': %w", templateName, err)
	}

	return renderedPrompt, nil
}

// Card represents a set of data to be injected into a prompt template
type Card struct {
	TemplateName string                 // The template to use
	Data         map[string]interface{} // Data to inject into the template
	Metadata     map[string]interface{} // Optional metadata about this card
}

// NewCard creates a new card for a specific template
func NewCard(templateName string, data map[string]interface{}) Card {
	return Card{
		TemplateName: templateName,
		Data:         data,
		Metadata:     make(map[string]interface{}),
	}
}

// Render applies the card data to its template and returns the rendered prompt
func (c *Card) Render(pm *PromptManager) (string, error) {
	if pm == nil {
		return "", errors.New("prompt manager is required to render a card")
	}

	return pm.RenderTemplate(c.TemplateName, c.Data)
}
