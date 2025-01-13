package main

import (
	"errors"

	"github.com/matigumma/memGo/utils"
)

type AzureOpenAIEmbedding struct {
	config BaseEmbedderConfig
}

func NewAzureOpenAIEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	utils.MapToStruct(config, &baseConfig)
	return &AzureOpenAIEmbedding{config: baseConfig}
}

func (a *AzureOpenAIEmbedding) Embed(text string) ([]float64, []float32, error) {
	return nil, nil, errors.New("AzureOpenAIEmbedding.Embed not implemented")
}
