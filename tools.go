package main

import "github.com/tmc/langchaingo/llms"

var ADD_MEMORY_TOOL = Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name:        "add_memory",
		Description: "Add a memory",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Data to add to memory",
				},
			},
			"required": []string{"data"},
		},
	},
}

var UPDATE_MEMORY_TOOL = Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name:        "update_memory",
		Description: "Update memory provided ID and data",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "memory_id of the memory to update",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Updated data for the memory",
				},
			},
			"required": []string{"memory_id", "data"},
		},
	},
}

var DELETE_MEMORY_TOOL = Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name:        "delete_memory",
		Description: "Delete memory by memory_id",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "memory_id of the memory to delete",
				},
			},
			"required": []string{"memory_id"},
		},
	},
}
