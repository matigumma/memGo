package tools

import (
	"github.com/matigumma/memGo/models"
	"github.com/tmc/langchaingo/llms"
)

var NO_OP_MEMORY_TOOL = models.Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name:        "no_op_memory",
		Description: "No operation on memory",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	},
}

var ADD_MEMORY_TOOL = models.Tool{
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

var UPDATE_MEMORY_TOOL = models.Tool{
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

var RESOLVE_MEMORY_CONFLICT_TOOL = models.Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name:        "resolve_memory_conflict",
		Description: "Resolve conflict between two memories using a specified strategy",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id_1": map[string]interface{}{
					"type":        "string",
					"description": "memory_id of the first memory",
				},
				"memory_id_2": map[string]interface{}{
					"type":        "string",
					"description": "memory_id of the second memory",
				},
				"strategy": map[string]interface{}{
					"type":        "string",
					"description": "Strategy to resolve the conflict (e.g., 'merge', 'prefer_first', 'prefer_second')",
				},
			},
			"required": []string{"memory_id_1", "memory_id_2", "strategy"},
		},
	},
}

var DELETE_MEMORY_TOOL = models.Tool{
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
