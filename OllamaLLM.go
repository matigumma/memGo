package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/llms"
)

type OllamaLLM struct {
	config BaseLlmConfig
}

func NewOllamaLLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	utils.MapToStruct(config, &baseConfig)
	return &OllamaLLM{config: baseConfig}
}

func (o *OllamaLLM) GenerateResponse(messages []llms.MessageContent, tools []models.Tool, jsonMode bool, toolChoice string) (interface{}, error) {
	return nil, errors.New("OllamaLLM.GenerateResponse not implemented")
}

func (o *OllamaLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("OllamaLLM.GenerateResponseWithoutTools not implemented")
}
