package main

import "errors"

type OpenAILLM struct {
	config BaseLlmConfig
}

func NewOpenAILLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	mapToStruct(config, &baseConfig)
	return &OpenAILLM{config: baseConfig}
}

func (o *OpenAILLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
	return nil, errors.New("OpenAILLM.GenerateResponse not implemented")
}

func (o *OpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("OpenAILLM.GenerateResponseWithoutTools not implemented")
}
