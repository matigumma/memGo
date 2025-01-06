package main

import "github.com/tmc/langchaingo/llms"

type Tool = llms.Tool

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
