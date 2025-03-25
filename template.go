package sapiens

import (
	"bytes"
	"fmt"
	"text/template"
)

// ApplyTemplate processes a template string with the provided data
func ApplyTemplate(templateStr string, data interface{}) (string, error) {
	// Create a new template
	tmpl, err := template.New("prompt").Funcs(template.FuncMap{
		"if": func(cond bool, v interface{}) interface{} {
			if cond {
				return v
			}
			return nil
		},
		"tone": func() string {
			return ""
		},
	}).Parse(templateStr)

	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with the data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
