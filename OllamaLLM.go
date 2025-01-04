package main

import "errors"

type OllamaLLM struct {
	config BaseLlmConfig
}

func NewOllamaLLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	mapToStruct(config, &baseConfig)
	return &OllamaLLM{config: baseConfig}
}

func (o *OllamaLLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
	return nil, errors.New("OllamaLLM.GenerateResponse not implemented")
}

func (o *OllamaLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("OllamaLLM.GenerateResponseWithoutTools not implemented")
}
