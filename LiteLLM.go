package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
)

type LiteLLM struct {
	config BaseLlmConfig
}

func NewLiteLLM(config BaseLlmConfig) *LiteLLM {
	return &LiteLLM{config: config}
}

func (l *LiteLLM) GenerateResponse(messages []map[string]string, tools []models.Tool) (map[string]interface{}, error) {
	return nil, errors.New("LiteLLM.GenerateResponse not implemented")
}

func (l *LiteLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("LiteLLM.GenerateResponseWithoutTools not implemented")
}
