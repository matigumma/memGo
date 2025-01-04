package main

import "errors"

type OpenAIEmbedding struct {
	config BaseEmbedderConfig
}

func NewOpenAIEmbedding(config BaseEmbedderConfig) *OpenAIEmbedding {
	return &OpenAIEmbedding{config: config}
}

func (o *OpenAIEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("OpenAIEmbedding.Embed not implemented")
}
