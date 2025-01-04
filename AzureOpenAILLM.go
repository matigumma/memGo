package main

import "errors"

type AzureOpenAILLM struct {
	config BaseLlmConfig
}

func NewAzureOpenAILLM(config BaseLlmConfig) *AzureOpenAILLM {
	return &AzureOpenAILLM{config: config}
}

func (a *AzureOpenAILLM) GenerateResponse(messages []map[string]string, tools []string) (map[string]interface{}, error) {
	return nil, errors.New("AzureOpenAILLM.GenerateResponse not implemented")
}

func (a *AzureOpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("AzureOpenAILLM.GenerateResponseWithoutTools not implemented")
}
