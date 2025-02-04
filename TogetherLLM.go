package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/llms"
)

type TogetherLLM struct {
	config BaseLlmConfig
}

func NewTogetherLLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	utils.MapToStruct(config, &baseConfig)
	return &TogetherLLM{config: baseConfig}
}

func (t *TogetherLLM) GenerateResponse(messages []llms.MessageContent, tools []models.Tool, jsonMode bool, toolChoice string) (interface{}, error) {
	return nil, errors.New("TogetherLLM.GenerateResponse not implemented")
}

func (t *TogetherLLM) GenerateResponseWithoutTools(messages []llms.MessageContent) (string, error) {
	return "", errors.New("TogetherLLM.GenerateResponseWithoutTools not implemented")
}
