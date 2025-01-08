package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIEmbedding struct {
	config *BaseEmbedderConfig
	client *openai.LLM
}

func NewOpenAIEmbedding(config map[string]interface{}) Embedder {
	baseConfig := BaseEmbedderConfig{}
	mapToStruct(config, &baseConfig)

	if baseConfig.Model == nil {
		defaultModel := "text-embedding-3-small"
		if baseConfig.Model == nil {
			baseConfig.Model = &defaultModel
		}
	}

	defaultDims := 1536
	if baseConfig.EmbeddingDims == nil {
		baseConfig.EmbeddingDims = &defaultDims
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" && baseConfig.APIKey != nil {
		apiKey = *baseConfig.APIKey
	}

	client, err := openai.New(openai.WithModel(*baseConfig.Model))
	if err != nil {
		fmt.Println("NewOpenAIEmbedding cliente fail.")
		log.Fatal(err)
	}

	// client := &OpenAIClient{APIKey: apiKey}
	return &OpenAIEmbedding{config: &baseConfig, client: client}
}

func (o *OpenAIEmbedding) Embed(text string) ([]float64, []float32, error) {
	text = strings.ReplaceAll(text, "\n", " ")

	embedding, err := o.client.CreateEmbedding(context.Background(), []string{text})
	if err != nil {
		return nil, nil, err
	}

	// Convert [][]float32 to []float64
	flatEmbedding := make([]float64, 0, len(embedding[0]))
	for _, value := range embedding[0] {
		flatEmbedding = append(flatEmbedding, float64(value))
	}

	return flatEmbedding, embedding[0], nil

	// return nil, errors.New("OpenAIEmbedding.Embed not implemented")
}
