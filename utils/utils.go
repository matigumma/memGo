package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matigumma/memGo/prompts"
	"github.com/tmc/langchaingo/llms"
)

func DebugPrint(message string, debug bool, gc *gin.Context) {
	if gc != nil {
		// gc.SSEvent("message", message)
		content := fmt.Sprintf("data: %s\n\n", message)
		_, err := gc.Writer.Write([]byte(content))
		if err != nil {
			fmt.Println("Error writing message:", err)
		}
		gc.Writer.Flush()
	}
	if debug {
		fmt.Println("DEBUG:::", message)
		// fmt.Println("") // print a separated line
	}
}

func MergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		merged[k] = v
	}
	return merged
}

func TrimMemorySuffix(s string) string {
	if len(s) > len("memory") && s[len(s)-len("memory"):] == "memory" {
		return s[:len(s)-len("memory")]
	}
	return s
}

// getUpdateMemoryPrompt formats the update memory prompt
func GetUpdateMemoryPrompt(existingMemories []map[string]interface{}, memory string, template string) string {
	var sb strings.Builder
	for _, mem := range existingMemories {
		if mem["score"] != nil {
			sb.WriteString(fmt.Sprintf("- ID: %s, Memory: %s, Score: %v\n", mem["id"], mem["memory"], mem["score"]))
		}
	}
	return fmt.Sprintf(template, sb.String(), memory)
}

// getUpdateMemoryMessages generates the messages for the LLM
func GetUpdateMemoryMessages(existingMemories []map[string]interface{}, memory string) llms.MessageContent {
	// prompt := GetUpdateMemoryPrompt(existingMemories, memory, prompts.GET_UPDATE_MEMORY_PROMPT_FUNCTION_CALLING)
	prompt := GetUpdateMemoryPrompt(existingMemories, memory, prompts.UPDATE_MEMORY_PROMPT)
	return llms.TextParts(llms.ChatMessageTypeHuman, prompt)
}

// --- Helper function to map a map to a struct ---
func MapToStruct(config map[string]interface{}, target interface{}) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}
	err = json.Unmarshal(configBytes, target)
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}
	return nil
}

/* ====== TOKEN COUNTER ====== */

// EstimateCost calculates the estimated cost in USD for LLM API usage
// based on input and output tokens. Currently only supports gpt-4o model.
/*
considerations:
gpt-4o: $2.50 / 1M input tokens
gpt-4o: $10.00 / 1M output tokens
gpt-4o-mini $0.150 / 1M input tokens
gpt-4o-mini $0.600 / 1M output token
text-embedding-3-small $0.020 / 1M tokens
*/
func EstimateCost(llmModel string, inputTokens, outputTokens int) float64 {
	switch llmModel {
	case "gpt-4o":
		return (float64(inputTokens) / 1000.0 * 0.0025) + (float64(outputTokens) / 1000.0 * 0.01)
	case "gpt-4o-mini":
		return (float64(inputTokens) / 1000.0 * 0.00015) + (float64(outputTokens) / 1000.0 * 0.0006)
	case "text-embedding-3-small":
		return float64(inputTokens) / 1000.0 * 0.00002 // $0.020 per 1M tokens
	default:
		return 0
	}
}
