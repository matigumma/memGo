package main

import (
	"fmt"
	"strings"
)

func mergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		merged[k] = v
	}
	return merged
}

func trimMemorySuffix(s string) string {
	if len(s) > len("memory") && s[len(s)-len("memory"):] == "memory" {
		return s[:len(s)-len("memory")]
	}
	return s
}

// getUpdateMemoryPrompt formats the update memory prompt
func getUpdateMemoryPrompt(existingMemories []map[string]interface{}, memory string, template string) string {
	var sb strings.Builder
	for _, mem := range existingMemories {
		sb.WriteString(fmt.Sprintf("- ID: %s, Memory: %s, Score: %v\n", mem["id"], mem["memory"], mem["score"]))
	}
	return fmt.Sprintf(template, sb.String(), memory)
}

// getUpdateMemoryMessages generates the messages for the LLM
func getUpdateMemoryMessages(existingMemories []map[string]interface{}, memory string) []map[string]string {
	prompt := getUpdateMemoryPrompt(existingMemories, memory, UPDATE_MEMORY_PROMPT)
	return []map[string]string{
		{"role": "user", "content": prompt},
	}
}
