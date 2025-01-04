package main

import "errors"

type HuggingFaceEmbedding struct {
	config BaseEmbedderConfig
}

func NewHuggingFaceEmbedding(config BaseEmbedderConfig) *HuggingFaceEmbedding {
	return &HuggingFaceEmbedding{config: config}
}

func (h *HuggingFaceEmbedding) Embed(text string) ([]float64, error) {
	return nil, errors.New("HuggingFaceEmbedding.Embed not implemented")
}
