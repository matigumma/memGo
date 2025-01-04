package main

import "errors"

type AzureOpenAIEmbedding struct {
	config BaseEmbedderConfig
}

func NewAzureOpenAIEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	mapToStruct(config, &baseConfig)
	return &AzureOpenAIEmbedding{config: baseConfig}
}

func (a *AzureOpenAIEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("AzureOpenAIEmbedding.Embed not implemented")
}
