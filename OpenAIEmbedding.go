package main

import "errors"

type OpenAIEmbedding struct {
	config BaseEmbedderConfig
}

func NewOpenAIEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	mapToStruct(config, &baseConfig)
	return &OpenAIEmbedding{config: baseConfig}
}

func (o *OpenAIEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("OpenAIEmbedding.Embed not implemented")
}
