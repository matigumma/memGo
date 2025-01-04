package main

import "errors"

type TogetherLLM struct {
	config BaseLlmConfig
}

func NewTogetherLLM(config BaseLlmConfig) *TogetherLLM {
	return &TogetherLLM{config: config}
}

func (t *TogetherLLM) GenerateResponse(messages []map[string]string, tools []string) (map[string]interface{}, error) {
	return nil, errors.New("TogetherLLM.GenerateResponse not implemented")
}

func (t *TogetherLLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
	return "", errors.New("TogetherLLM.GenerateResponseWithoutTools not implemented")
}
