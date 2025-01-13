package main

import (
	"errors"

	"github.com/matigumma/memGo/utils"
)

type HuggingFaceEmbedding struct {
	config BaseEmbedderConfig
}

func NewHuggingFaceEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	utils.MapToStruct(config, &baseConfig)
	return &HuggingFaceEmbedding{config: baseConfig}
}

func (h *HuggingFaceEmbedding) Embed(text string) ([]float64, []float32, error) {
	return nil, nil, errors.New("HuggingFaceEmbedding.Embed not implemented")
}
