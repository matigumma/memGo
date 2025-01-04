package main

import "errors"

type OllamaEmbedding struct {
	config BaseEmbedderConfig
}

func NewOllamaEmbedding(config BaseEmbedderConfig) *OllamaEmbedding {
	return &OllamaEmbedding{config: config}
}

func (o *OllamaEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("OllamaEmbedding.Embed not implemented")
}
