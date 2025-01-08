package main

import "errors"

type OllamaEmbedding struct {
	config BaseEmbedderConfig
}

func NewOllamaEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	mapToStruct(config, &baseConfig)
	return &OllamaEmbedding{config: baseConfig}
}

func (o *OllamaEmbedding) Embed(text string) ([]float64, []float32, error) {
	return nil, nil, errors.New("OllamaEmbedding.Embed not implemented")
}
