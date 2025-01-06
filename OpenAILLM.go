package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAILLM struct {
	config *BaseLlmConfig
	client *openai.LLM
}

func NewOpenAILLM(config map[string]interface{}) *OpenAILLM {
	baseConfig := BaseLlmConfig{}
	mapToStruct(config, &baseConfig)

	if baseConfig.Model == nil {
		defaultModel := "gpt-4o-mini"
		if baseConfig.Model == nil {
			baseConfig.Model = &defaultModel
		}
	}

	var client *openai.LLM
	// if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
	// 	client = NewOpenAIClient(apiKey, baseConfig.OpenrouterBaseURL)
	// } else {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = *baseConfig.APIKey
	}
	client, err := openai.New(openai.WithModel(*baseConfig.Model))
	if err != nil {
		fmt.Println("NewOpenAILLM client fail.")
		log.Fatal(err)
	}
	// }

	return &OpenAILLM{config: &baseConfig, client: client}
}

func (o *OpenAILLM) parseResponse(response *llm.response, tools []Tool) (interface{}, error) {
	if tools != nil {
		processedResponse := map[string]interface{}{
			"content":    response.Choices[0].Message.Content,
			"tool_calls": []interface{}{},
		}

		if response.Choices[0].Message.ToolCalls != nil {
			for _, toolCall := range response.Choices[0].Message.ToolCalls {
				arguments := map[string]interface{}{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
					return nil, err
				}
				processedResponse["tool_calls"] = append(processedResponse["tool_calls"].([]interface{}), map[string]interface{}{
					"name":      toolCall.Function.Name,
					"arguments": arguments,
				})
			}
		}

		return processedResponse, nil
	} else {
		return response.Choices[0].Message.Content, nil
	}
}

/*

func (o *OpenAILLM) GenerateResponse(messages []map[string]string, responseFormat interface{}, tools []Tool, toolChoice string) (interface{}, error) {
	params := map[string]interface{}{
		"model":       o.config.Model,
		"messages":    messages,
		"temperature": o.config.Temperature,
		"max_tokens":  o.config.MaxTokens,
		"top_p":       o.config.TopP,
	}

	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		openrouterParams := map[string]interface{}{}
		if len(o.config.Models) > 0 {
			openrouterParams["models"] = o.config.Models
			openrouterParams["route"] = o.config.Route
			delete(params, "model")
		}

		if o.config.SiteURL != "" && o.config.AppName != "" {
			extraHeaders := map[string]string{
				"HTTP-Referer": o.config.SiteURL,
				"X-Title":      o.config.AppName,
			}
			openrouterParams["extra_headers"] = extraHeaders
		}

		for k, v := range openrouterParams {
			params[k] = v
		}
	}

	if responseFormat != nil {
		params["response_format"] = responseFormat
	}
	if tools != nil {
		params["tools"] = tools
		params["tool_choice"] = toolChoice
	}

	response, err := o.client.ChatCompletionsCreate(params)
	if err != nil {
		return nil, err
	}

	return o.parseResponse(response, tools)
}

*/

// func (o *OpenAILLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
// 	return nil, errors.New("OpenAILLM.GenerateResponse not implemented")
// }

// func (o *OpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
// 	return "", errors.New("OpenAILLM.GenerateResponseWithoutTools not implemented")
// }
