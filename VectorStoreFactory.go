package main

import "fmt"

// VectorStoreFactory - Corresponds to the Python VectorStoreFactory class
type VectorStoreFactory struct{}

// Create - Creates a VectorStore instance based on the provider name and configuration
func (vsf VectorStoreFactory) Create(providerName string, config map[string]interface{}) (VectorStore, error) {
	switch providerName {
	case "qdrant":
		return NewQdrant(config), nil
	case "chroma":
		return NewChromaDB(config), nil
	case "pgvector":
		return NewPGVector(config), nil
	default:
		return nil, fmt.Errorf("unsupported VectorStore provider: %s", providerName)
	}
}
