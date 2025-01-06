package main

import "fmt"

type LlmFactory struct{}

func (lf LlmFactory) Create(providerName string, config map[string]interface{}) (LLM, error) {
	switch providerName {
	case "ollama":
		return NewOllamaLLM(config), nil
	case "openai":
		return NewOpenAILLM(config), nil
	case "together":
		return NewTogetherLLM(config), nil
	case "azure_openai":
		return NewAzureOpenAILLM(config), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", providerName)
	}
}
