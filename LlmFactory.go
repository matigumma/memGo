package main

import "fmt"

// LlmFactory - Corresponds to the Python LlmFactory class
type LlmFactory struct{}

// Create - Creates an LLM instance based on the provider name and configuration
func (lf LlmFactory) Create(providerName string, config map[string]interface{}) (LLM, error) {
	switch providerName {
	case "ollama":
		cfg := BaseLlmConfig{}
		// Use a library like `mapstructure` for robust map to struct conversion if needed
		// For simplicity, manually assign if the map structure is simple and known
		if val, ok := config["model"].(string); ok {
			cfg.Model = &val
		}
		// ... assign other fields ...
		return NewOllamaLLM(cfg), nil
	case "openai":
		cfg := BaseLlmConfig{}
		// ... assign fields ...
		return NewOpenAILLM(cfg), nil
	case "groq":
		return NewGroqLLM(BaseLlmConfig{}), nil
	case "together":
		return NewTogetherLLM(BaseLlmConfig{}), nil
	case "aws_bedrock":
		return NewAWSBedrockLLM(BaseLlmConfig{}), nil
	case "litellm":
		return NewLiteLLM(BaseLlmConfig{}), nil
	case "azure_openai":
		return NewAzureOpenAILLM(BaseLlmConfig{}), nil
	default:
		return nil, fmt.Errorf("unsupported Llm provider: %s", providerName)
	}
}
