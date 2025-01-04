package main

import "errors"

type AzureOpenAIEmbedding struct {
	config BaseEmbedderConfig
}

func NewAzureOpenAIEmbedding(config BaseEmbedderConfig) *AzureOpenAIEmbedding {
	return &AzureOpenAIEmbedding{config: config}
}

func (a *AzureOpenAIEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("AzureOpenAIEmbedding.Embed not implemented")
}
