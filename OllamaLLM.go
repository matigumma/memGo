package main

import "errors"

type OllamaLLM struct {
	config BaseLlmConfig
}

func NewOllamaLLM(config BaseLlmConfig) *OllamaLLM {
	return &OllamaLLM{config: config}
}

func (o *OllamaLLM) GenerateResponse(messages []map[string]string, tools []string) (map[string]interface{}, error) {
	return nil, errors.New("OllamaLLM.GenerateResponse not implemented")
}

func (o *OllamaLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("OllamaLLM.GenerateResponseWithoutTools not implemented")
}
