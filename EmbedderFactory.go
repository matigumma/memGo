package main

import "fmt"

// EmbedderFactory - Corresponds to the Python EmbedderFactory class
type EmbedderFactory struct{}

// Create - Creates an Embedder instance based on the provider name and configuration
// func (ef EmbedderFactory) Create(providerName string, config map[string]interface{}) (Embedder, error) {
// 	switch providerName {
// 	case "openai":
// 		cfg := BaseEmbedderConfig{}
// 		// ... assign fields ...
// 		return NewOpenAIEmbedding(cfg), nil
// 	case "ollama":
// 		cfg := BaseEmbedderConfig{}
// 		// ... assign fields ...
// 		return NewOllamaEmbedding(cfg), nil
// 	case "huggingface":
// 		return NewHuggingFaceEmbedding(BaseEmbedderConfig{}), nil
// 	case "azure_openai":
// 		return NewAzureOpenAIEmbedding(BaseEmbedderConfig{}), nil
// 	default:
// 		return nil, fmt.Errorf("unsupported Embedder provider: %s", providerName)
// 	}
// }

func (ef EmbedderFactory) Create(providerName string, config map[string]interface{}) (Embedder, error) {
	switch providerName {
	case "openai":
		return NewOpenAIEmbedding(config), nil
	case "ollama":
		return NewOllamaEmbedding(config), nil
	case "huggingface":
		return NewHuggingFaceEmbedding(config), nil
	case "azure_openai":
		return NewAzureOpenAIEmbedding(config), nil
	default:
		return nil, fmt.Errorf("unsupported Embedder provider: %s", providerName)
	}
}
