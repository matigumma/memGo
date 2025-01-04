package main

import "errors"

type TogetherLLM struct {
	config BaseLlmConfig
}

func NewTogetherLLM(config map[string]interface{}) LLM {
	baseConfig := BaseLlmConfig{}
	mapToStruct(config, &baseConfig)
	return &TogetherLLM{config: baseConfig}
}

func (t *TogetherLLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
	return nil, errors.New("TogetherLLM.GenerateResponse not implemented")
}

func (t *TogetherLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("TogetherLLM.GenerateResponseWithoutTools not implemented")
}
