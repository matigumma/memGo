package main

import "fmt"

// BaseEmbedderConfig - Corresponds to the Python BaseEmbedderConfig class
type BaseEmbedderConfig struct {
	Model         *string                `json:"model,omitempty"`
	APIKey        *string                `json:"api_key,omitempty"`
	EmbeddingDims *int                   `json:"embedding_dims,omitempty"`
	OllamaBaseURL *string                `json:"ollama_base_url,omitempty"`
	ModelKwargs   map[string]interface{} `json:"model_kwargs,omitempty"`
}

// EmbedderConfig - Corresponds to the Python EmbedderConfig class
type EmbedderConfig struct {
	Provider string                 `json:"provider" default:"openai"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ValidateConfig validates the EmbedderConfig
func (ec *EmbedderConfig) ValidateConfig() error {
	switch ec.Provider {
	case "openai", "ollama", "huggingface", "azure_openai":
		return nil
	default:
		return fmt.Errorf("unsupported embedding provider: %s", ec.Provider)
	}
}
