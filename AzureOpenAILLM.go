package main

import (
	"errors"

	"github.com/tmc/langchaingo/llms"
)

type AzureOpenAILLM struct {
	config BaseLlmConfig
}

func NewAzureOpenAILLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	mapToStruct(config, &baseConfig)
	return &AzureOpenAILLM{config: baseConfig}
}

func (a *AzureOpenAILLM) GenerateResponse(messages []llms.MessageContent, tools []Tool, jsonMode bool, toolChoice string) (interface{}, error) {
	return nil, errors.New("AzureOpenAILLM.GenerateResponse not implemented")
}

func (a *AzureOpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("AzureOpenAILLM.GenerateResponseWithoutTools not implemented")
}
