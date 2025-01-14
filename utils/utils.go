package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matigumma/memGo/prompts"
	"github.com/tmc/langchaingo/llms"
)

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
