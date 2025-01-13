package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/matigumma/memGo/models"
	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAILLM struct {
	config *BaseLlmConfig
	client *openai.LLM
}

func NewOpenAILLM(config map[string]interface{}) *OpenAILLM {
	baseConfig := BaseLlmConfig{}
	utils.MapToStruct(config, &baseConfig)

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

// parseResponse takes a ContentResponse and an optional list of Tools and returns a response
// object that is either a string or a map with "content" and "tool_calls" keys.
func (o *OpenAILLM) parseResponse(
	response *llms.ContentResponse, // The raw response from API
	tools []models.Tool, // List of tools
) (interface{}, error) {
	// Check if tools are provided
	if tools != nil {
		// Initialize a map to hold the processed response with "content" and "tool_calls" keys
		processedResponse := map[string]interface{}{
			"content":    response.Choices[0].Content, // Set the content from the first choice
			"tool_calls": []interface{}{},             // Initialize an empty slice for tool calls
		}

		// Check if there are any tool calls in the response
		if response.Choices[0].ToolCalls != nil {
			// Iterate over each tool call
			for _, toolCall := range response.Choices[0].ToolCalls {
				// Initialize a map to hold the arguments for the tool call
				arguments := map[string]interface{}{}
				// Unmarshal the JSON arguments into the map
				if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &arguments); err != nil {
					return nil, err // Return an error if unmarshalling fails
				}
				// Append the tool call details to the "tool_calls" slice in the processed response
				processedResponse["tool_calls"] = append(processedResponse["tool_calls"].([]interface{}), map[string]interface{}{
					"name":      toolCall.FunctionCall.Name, // Set the tool call name
					"arguments": arguments,                  // Set the tool call arguments
				})
			}
		}

		// Return the processed response map
		return processedResponse, nil
	} else {
		// If no tools are provided, return just the content from the first choice
		return response.Choices[0].Content, nil
	}
}

// generate response
func (o *OpenAILLM) GenerateResponse(
	messages []llms.MessageContent, // List of messages
	tools []models.Tool, // List of tools
	jsonMode bool, // Flag to indicate JSON mode
	toolChoice string, // Tool choice
) (interface{}, error) {

	options := llms.CallOptions{}
	options.Model = *o.config.Model
	options.Temperature = o.config.Temperature
	options.MaxTokens = o.config.MaxTokens

	// params := map[string]interface{}{
	// 	"model":       o.config.Model,
	// 	"messages":    messages,
	// 	"temperature": o.config.Temperature,
	// 	"max_tokens":  o.config.MaxTokens,
	// 	"top_p":       o.config.TopP,
	// }

	// if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
	// 	openrouterParams := map[string]interface{}{}
	// 	if len(o.config.Models) > 0 {
	// 		openrouterParams["models"] = o.config.Models
	// 		openrouterParams["route"] = o.config.Route
	// 		delete(params, "model")
	// 	}

	// 	if o.config.SiteURL != "" && o.config.AppName != "" {
	// 		extraHeaders := map[string]string{
	// 			"HTTP-Referer": o.config.SiteURL,
	// 			"X-Title":      o.config.AppName,
	// 		}
	// 		openrouterParams["extra_headers"] = extraHeaders
	// 	}

	// 	for k, v := range openrouterParams {
	// 		params[k] = v
	// 	}
	// }

	// if tools != nil {
	// 	params["tools"] = tools
	// 	params["tool_choice"] = o.config.toolChoice
	// }

	if jsonMode {
		options.JSONMode = true
	}

	if tools != nil {
		options.Tools = tools
		if toolChoice != "" {
			options.ToolChoice = toolChoice
		} else {
			options.ToolChoice = "auto"
		}
	}

	// response, err := o.client.ChatCompletionsCreate(params)
	response, err := o.client.GenerateContent(context.Background(), messages, llms.WithOptions(options))
	if err != nil {
		return nil, err
	}

	return o.parseResponse(response, tools)
}

/*

 */

// func (o *OpenAILLM) GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error) {
// 	return nil, errors.New("OpenAILLM.GenerateResponse not implemented")
// }

// func (o *OpenAILLM) GenerateResponseWithoutTools(messages []map[string]string) (string, error) {
// 	return "", errors.New("OpenAILLM.GenerateResponseWithoutTools not implemented")
// }
