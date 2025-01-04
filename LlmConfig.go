package main

import "fmt"

// BaseLlmConfig - Corresponds to the Python BaseLlmConfig class
type BaseLlmConfig struct {
	Model             *string  `json:"model,omitempty"`
	Temperature       float64  `json:"temperature,omitempty" default:"0"`
	APIKey            *string  `json:"api_key,omitempty"`
	MaxTokens         int      `json:"max_tokens,omitempty" default:"3000"`
	TopP              float64  `json:"top_p,omitempty" default:"0"`
	TopK              int      `json:"top_k,omitempty" default:"1"`
	Models            []string `json:"models,omitempty"`
	Route             *string  `json:"route,omitempty" default:"fallback"`
	OpenrouterBaseURL *string  `json:"openrouter_base_url,omitempty" default:"https://openrouter.ai/api/v1"`
	SiteURL           *string  `json:"site_url,omitempty"`
	AppName           *string  `json:"app_name,omitempty"`
	OllamaBaseURL     *string  `json:"ollama_base_url,omitempty"`
}

// LlmConfig - Corresponds to the Python LlmConfig class
type LlmConfig struct {
	Provider string                 `json:"provider" default:"openai"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ValidateConfig validates the LlmConfig
func (lc *LlmConfig) ValidateConfig() error {
	switch lc.Provider {
	case "openai", "ollama", "groq", "together", "aws_bedrock", "litellm", "azure_openai":
		return nil
	default:
		return fmt.Errorf("unsupported LLM provider: %s", lc.Provider)
	}
}
