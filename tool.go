package sapiens

// NewTool creates a new tool
func NewTool(name string, description string, requiredTools []Tool) Tool {
	return Tool{
		Name:          name,
		Description:   description,
		RequiredTools: requiredTools,
	}
}

// SetCost sets the cost of using this tool
func (t *Tool) SetCost(cost float64) {
	t.Cost = cost
}

// AddInputSchema adds an input schema to the tool
func (t *Tool) AddInputSchema(inputFormat *Schema) {
	t.InputSchema = inputFormat
}

// AddOutputSchema adds an output schema to the tool
func (t *Tool) AddOutputSchema(outputFormat *Schema) {
	t.OutputSchema = outputFormat
}

// AddRequiredTool adds a required tool dependency
func (t *Tool) AddRequiredTool(tool Tool) {
	t.RequiredTools = append(t.RequiredTools, tool)
}

// SetDescription sets the description of the tool
func (t *Tool) SetDescription(description string) {
	t.Description = description
}
