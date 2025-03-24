package sapiens

func NewTool(name string, description string, requiredTools []Tool) Tool {
	return Tool{
		Name:          name,
		Description:   description,
		RequiredTools: requiredTools,
	}
}

func (t *Tool) SetCost(cost float64) {
	t.Cost = cost
}

func (t *Tool) AddInputSchema(inputFormat *Schema) {
	t.InputSchema = inputFormat
}

func (t *Tool) AddOutputSchema(outputFormat *Schema) {
	t.OutputSchema = outputFormat
}

func (t *Tool) AddRequiredTool(tool Tool) {
	t.RequiredTools = append(t.RequiredTools, tool)
}

func (t *Tool) SetDescription(description string) {
	t.Description = description
}
