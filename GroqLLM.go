package main

import (
	"errors"

	"github.com/matigumma/memGo/models"
)

type GroqLLM struct {
	config BaseLlmConfig
}

func NewGroqLLM(config BaseLlmConfig) *GroqLLM {
	return &GroqLLM{config: config}
}

func (g *GroqLLM) GenerateResponse(messages []map[string]string, tools []models.Tool) (map[string]interface{}, error) {
	return nil, errors.New("GroqLLM.GenerateResponse not implemented")
}

func (g *GroqLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("GroqLLM.GenerateResponseWithoutTools not implemented")
}
