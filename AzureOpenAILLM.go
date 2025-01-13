package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/llms"
)

type AzureOpenAILLM struct {
	config BaseLlmConfig
}

func NewAzureOpenAILLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	utils.MapToStruct(config, &baseConfig)
	return &AzureOpenAILLM{config: baseConfig}
}

func (a *AzureOpenAILLM) GenerateResponse(messages []llms.MessageContent, tools []models.Tool, jsonMode bool, toolChoice string) (interface{}, error) {
	return nil, errors.New("AzureOpenAILLM.GenerateResponse not implemented")
}

func (a *AzureOpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("AzureOpenAILLM.GenerateResponseWithoutTools not implemented")
}
