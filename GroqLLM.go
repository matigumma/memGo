package main

import "errors"

type GroqLLM struct {
	config BaseLlmConfig
}

func NewGroqLLM(config BaseLlmConfig) *GroqLLM {
	return &GroqLLM{config: config}
}

func (g *GroqLLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
	return nil, errors.New("GroqLLM.GenerateResponse not implemented")
}

func (g *GroqLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("GroqLLM.GenerateResponseWithoutTools not implemented")
}
