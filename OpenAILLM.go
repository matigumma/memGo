package main

import "errors"

type OpenAILLM struct {
	config BaseLlmConfig
}

func NewOpenAILLM(config BaseLlmConfig) *OpenAILLM {
	return &OpenAILLM{config: config}
}

func (o *OpenAILLM) GenerateResponse(messages []map[string]string, tools []string) (map[string]interface{}, error) {
	return nil, errors.New("OpenAILLM.GenerateResponse not implemented")
}

func (o *OpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("OpenAILLM.GenerateResponseWithoutTools not implemented")
}
