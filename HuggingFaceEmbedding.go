package main

import "errors"

type HuggingFaceEmbedding struct {
	config BaseEmbedderConfig
}

func NewHuggingFaceEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	mapToStruct(config, &baseConfig)
	return &HuggingFaceEmbedding{config: baseConfig}
}

func (h *HuggingFaceEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("HuggingFaceEmbedding.Embed not implemented")
}
