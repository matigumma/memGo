package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
)

type AWSBedrockLLM struct {
	config BaseLlmConfig
}

func NewAWSBedrockLLM(config BaseLlmConfig) *AWSBedrockLLM {
	return &AWSBedrockLLM{config: config}
}

func (a *AWSBedrockLLM) GenerateResponse(messages []map[string]string, tools []models.Tool) (map[string]interface{}, error) {
	return nil, errors.New("AWSBedrockLLM.GenerateResponse not implemented")
}

func (a *AWSBedrockLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("AWSBedrockLLM.GenerateResponseWithoutTools not implemented")
}
