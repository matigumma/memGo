package main

type Tool struct {
	// Type is the type of the tool.
	Type string `json:"type"`
	// Function is the function to call.
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters,omitempty"`
}

// MemoryItem - Corresponds to the Python MemoryItem class
type MemoryItem struct {
	ID        string                 `json:"id"`
	Memory    string                 `json:"memory"`
	Hash      *string                `json:"hash,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Score     *float64               `json:"score,omitempty"`
	CreatedAt *string                `json:"created_at,omitempty"`
	UpdatedAt *string                `json:"updated_at,omitempty"`
}
