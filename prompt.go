package sapiens

import (
	"bytes"
	"text/template"
)

func NewSystemPrompt(prompt string, version string) Prompt {
	return Prompt{
		Prompt:  prompt,
		Version: version,
	}
}

func (s *Prompt) SetEnhanced(enhanced Prompt) {
	s.Enhanced = append(s.Enhanced, enhanced)
}

func (s *Prompt) GetLatestEnhanced() Prompt {
	if len(s.Enhanced) == 0 {
		return Prompt{}
	}
	return s.Enhanced[len(s.Enhanced)-1]
}

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
