package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/matigumma/memGo/models"
	p "github.com/matigumma/memGo/prompts"
	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type Chain struct {
	debug bool
}

func NewChain(debug bool) *Chain {
	return &Chain{
		debug: debug,
	}
}

func (c *Chain) Cb(ctx context.Context, m map[string]any) {
	c.debugPrint("callback: " + fmt.Sprintf("%v", m))
}

func (c *Chain) MEMORY_DEDUCTION(data string) (map[string]interface{}, error) {
	c.debugPrint("Chain.MEMORY_DEDUCTION")
	st := time.Now()

	/* ====== SETTINGS ====== */
	// model := "gpt-3.5-turbo" // este funciona ahi nomas,,, se queda medio corto
	model := "gpt-4o-mini" // funciona bien con el ultimo prompt
	// model := "gpt-4o"
	/* ====== LLM INSTANCE ====== */

	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("error creating LLM: %w", err)
	}

	/* ====== PROMPT ====== */
	prompt := prompts.NewPromptTemplate(
		p.MEMORY_DEDUCTION_PROMPT_SPA,
		[]string{"conversation"},
	)

	/* ====== DATA FORMAT ====== */
	messages := []llms.MessageContent{}
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, data))

	/* ====== CHAIN ====== */
	ctx := context.Background()
	llmChain := chains.NewLLMChain(llm, prompt)
	out, err := chains.Call(ctx, llmChain, map[string]any{
		"conversation": messages,
	})
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	/* ====== DEBUG ====== */
	c.debugPrint("Using model: " + model)
	c.debugPrint("Output from LLM: " + fmt.Sprintf("%v", out))
	c.debugPrint("Token Cost: " + fmt.Sprintf("%.6f", utils.EstimateCost(model, out["input_tokens"].(int), out["output_tokens"].(int))))

	/* ====== OUTPUT FORMAT ====== */
	parsedOutput, ok := out["text"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse output text")
	}

	// remove possible trailing ```json and ``` from llm output text
	parsedOutput = strings.Trim(parsedOutput, "```json")
	parsedOutput = strings.Trim(parsedOutput, "`")

	var result map[string]interface{}
	// stores the parsedOutput in the result value pointer
	err = json.Unmarshal([]byte(parsedOutput), &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	/* ====== OUTPUT SCHEMA ====== */
	/*
		{
			"relevant_facts": {
				"type": "array",
				"items": {
					"type": "string"
				}
			},
			"metadata": {
				"type": "object",
				"properties": {
					"scope": {
						"type": "string"
					},
					"sentiment": {
						"type": "string",
						"description": "The overall sentiment of the text, e.g., positive, negative, neutral"
					}
					"related_entities": {
						"type": "array",
						"items": {
							"type": "string"
						}
					},
					"related_events": {
						"type": "array",
						"items": {
							"type": "string"
						}
					},
					"tags": {
						"type": "array",
						"items": {
							"type": "string"
						}
					}
				}
			}
		}
	*/

	elapsed := time.Since(st)
	c.debugPrint("Chain.MEMORY_DEDUCTION took: " + elapsed.String())
	return result, nil
}

func (c *Chain) MEMORY_UPDATER(existingMemories []models.MemoryItem, relevantFacts []interface{}) (*llms.ContentChoice, error) {
	c.debugPrint("Chain.MEMORY_UPDATER")
	st := time.Now()

	/* ====== SETTINGS ====== */
	ctx := context.Background()

	// model := "gpt-3.5-turbo" //
	model := "gpt-4o-mini" //
	// model := "gpt-4o"

	/* ====== LLM INSTANCE ====== */
	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("error creating LLM: %w", err)
	}

	/* ====== PROMPT ====== */

	/* ====== DATA FORMAT ====== */
	serializedExistingMemories := make([]map[string]interface{}, len(existingMemories))
	for i, item := range existingMemories {
		if item.Score != nil {
			serializedItem := map[string]interface{}{
				"real_id": item.ID,
				"id":      i,
				"memory":  item.Memory,
				"score":   *item.Score,
			}
			serializedExistingMemories[i] = serializedItem
			c.debugPrint("SRZd memory: " + fmt.Sprintf("%+v", serializedItem))
		}
	}

	var newFacts []string
	var relevantFactsText string
	for i, fact := range relevantFacts {
		if factStr, ok := fact.(string); ok {
			relevantFactsText += fmt.Sprintf("%d - %s\n ", i+1, factStr)
			c.debugPrint("RLVNt fact: " + factStr)
		} else {
			c.debugPrint(fmt.Sprintf("Skipping non-string fact: %v", fact))
		}
	}

	// tools := []models.Tool{tools.ADD_MEMORY_TOOL, tools.UPDATE_MEMORY_TOOL, tools.DELETE_MEMORY_TOOL}

	Relevancia := chains.NewLLMChain(llm, prompts.NewPromptTemplate(`You are tasked with determining the relevance of newly retrieved facts to the existing memory. Compare the new facts with each memory entry and assign one of the following statuses:
- MATCH: The fact is identical or highly similar to the memory.
- EXTEND: The fact provides additional information about an existing memory.
- CONFLICT: The fact directly contradicts an existing memory.
- NEW: The fact is unrelated to existing memories.

Existing Memory:
{{.existing_memories}}

New Retrieved Facts:
{{.relevantFactsText}}

Return a JSON object where each fact is mapped to one of the above statuses, with a brief explanation of the reasoning for the status assignment.
`, []string{
		"existing_memories",
		"relevantFactsText",
	},
	))

	/* output example:
	{
		"1": {
			"status": "MATCH",
			"reason": "The fact is identical to the existing memory entry with id:20, which states that a contact was received from a potential client interested in implementing a chatbot."
		},
		"2": {
			"status": "EXTEND",
			"reason": "This fact provides additional information about the specific functionality of the chatbot, which is to analyze WhatsApp conversations, extending the existing memory about the client's needs."
		},
		"3": {
			"status": "NEW",
			"reason": "This fact introduces new information regarding the client's economic situation and the developer's considerations, which is not covered in the existing memories."
		},
	}
	*/
	RelevanciaResponse, err := chains.Call(ctx, Relevancia, map[string]interface{}{
		"existing_memories": serializedExistingMemories,
		"relevantFactsText": relevantFactsText,
	})

	if err != nil {
		log.Panic(err)
	}

	// Iterate through the RelevanciaResponse to filter facts
	newrelevantFactsText := ""
	newFacts = make([]string, 0)
	textResponse, ok := RelevanciaResponse["text"].(string)
	if !ok {
		log.Panic("RelevanciaResponse['text'] is not a map")
	}

	textResponse = strings.Trim(textResponse, "```json")
	textResponse = strings.Trim(textResponse, "```")

	var textResponseResult map[string]interface{}
	err = json.Unmarshal([]byte(textResponse), &textResponseResult)
	if err != nil {
		log.Panic("Error parsing JSON: ", err)
	}

	c.debugPrint("RelevanciaResponse: " + fmt.Sprintf("%+v", textResponseResult))

	idsToDeleteFromResponseResult := make([]string, 0)
	for key, value := range textResponseResult {

		statusInfo, ok := value.(map[string]interface{})
		if !ok {
			log.Printf("Skipping non-map value for key %s: %v", key, value)
			continue
		}
		status, ok := statusInfo["status"].(string)
		if !ok {
			log.Printf("Skipping entry for key %s: status is not a string", key)
			continue
		}

		if status == "MATCH" {
			// los que tienen MATCH son memorias que no necesito actualizar

			for i, fact := range relevantFacts {
				if strconv.Itoa(i+1) == key {
					c.debugPrint("MATCH: " + fmt.Sprintf("%+v", fact))
					if factStr, ok := fact.(string); ok {
						newrelevantFactsText += fmt.Sprintf("%s - %s\n ", key, factStr)
						idsToDeleteFromResponseResult = append(idsToDeleteFromResponseResult, key)
					} else {
						// showld i panic?
						log.Printf("Skipping non-string fact: %v", fact)
					}
				}
			}

		} else if status == "NEW" {
			// los que tienen NEW son memorias nuevas que necesito actualizar

			for i, fact := range relevantFacts {
				if strconv.Itoa(i+1) == key {
					c.debugPrint("NEW: " + fmt.Sprintf("%+v", fact))
					if factStr, ok := fact.(string); ok {
						newFacts = append(newFacts, factStr) // FACTS to add into memory
						idsToDeleteFromResponseResult = append(idsToDeleteFromResponseResult, key)
					} else {
						// showld i panic?
						log.Printf("Skipping non-string fact: %v", fact)
					}
				}
			}
		} else if status == "CONFLICT" {
			for i, fact := range relevantFacts {
				if strconv.Itoa(i+1) == key {
					c.debugPrint("CONFLICT: " + fmt.Sprintf("%+v", fact))
					if factStr, ok := fact.(string); ok {
						textResponseResult[key] = map[string]interface{}{
							"status": "CONFLICT",
							"fact":   factStr,
						}
					} else {
						// showld i panic?
						log.Printf("Skipping non-string fact: %v", fact)
					}
				}
			}
		} else if status == "EXTEND" {
			for i, fact := range relevantFacts {
				if strconv.Itoa(i+1) == key {
					c.debugPrint("EXTEND: " + fmt.Sprintf("%+v", fact))
					if factStr, ok := fact.(string); ok {
						textResponseResult[key] = map[string]interface{}{
							"status": "EXTEND",
							"reason": factStr,
						}
					} else {
						// showld i panic?
						log.Printf("Skipping non-string fact: %v", fact)
					}
				}
			}
		}
	}

	for _, id := range idsToDeleteFromResponseResult {
		delete(textResponseResult, id)
	}

	// c.debugPrint("NEW FACTS: " + strings.Join(newFacts, "\n"))
	// c.debugPrint("NEW RELEVANT FACTS: " + newrelevantFactsText)
	relevantFactsText = newrelevantFactsText // lo piso...

	c.debugPrint("FINAL RelevanciaResponse: " + fmt.Sprintf("%+v", textResponseResult))

	Conflictos := chains.NewLLMChain(llm, prompts.NewPromptTemplate(`You are analyzing conflicting facts between the existing memory and newly retrieved facts. Your task is to resolve these conflicts based on the following criteria:
1. If the new fact is more detailed or recent, replace the old memory.
2. If the old memory is more accurate or detailed, retain it and discard the new fact.

Old Memory:
{{.existing_memories}}

Conflicting Pairs:
{{.conflicts}}

Return a JSON object indicating which memory to retain and which to discard. Include reasoning for each decision.
`, []string{
		"conflicts",
	},
	))

	ConflictosResponse, err := chains.Call(ctx, Conflictos, map[string]interface{}{
		"conflicts":         fmt.Sprintf("%+v", relevantFactsText),
		"existing_memories": serializedExistingMemories,
	})

	if err != nil {
		log.Panic(err)
	}

	c.debugPrint("CONFLICTOSRESPONSE::: " + fmt.Sprintf("%+v", ConflictosResponse))

	// return nil, fmt.Errorf("asd")

	Actualizar := chains.NewLLMChain(llm, prompts.NewPromptTemplate(`You are tasked with identifying updates to the memory based on the retrieved facts marked as EXTEND. Your goal is to:
1. Merge the new fact into the corresponding memory entry, preserving its original ID.
2. Ensure the updated memory is concise and informative.

Memory Entries to Update:
{{.existing_memories}}

New Retrieved Facts to Extend:
{{.conflicts}}

Return a JSON object where each updated memory includes its ID, the updated text, and the reasoning behind the update.
`, []string{
		"existing_memories",
		"conflicts",
	},
	))

	ActualizarResponse, err := chains.Call(ctx, Actualizar, map[string]interface{}{
		"existing_memories": serializedExistingMemories,
		"conflicts":         ConflictosResponse,
	})

	if err != nil {
		log.Panic(err)
	}

	c.debugPrint("ActualizarResponse: " + fmt.Sprintf("%+v", ActualizarResponse))

	Identificar := chains.NewLLMChain(llm, prompts.NewPromptTemplate(`You are tasked with identifying new facts that need to be added to the memory. Each new fact should receive a unique ID.

Existing Memory:
{{.existing_memories}}

New Retrieved Facts to Add:
{{.conflicts}}

Return a JSON object with the new facts added, including their assigned IDs.

`, []string{
		"existing_memories",
		"conflicts",
	},
	))

	IdentificarResponse, err := chains.Call(ctx, Identificar, map[string]interface{}{
		"existing_memories": serializedExistingMemories,
		"conflicts":         ActualizarResponse,
	})

	if err != nil {
		log.Panic(err)
	}

	c.debugPrint("IdentificarResponse: " + fmt.Sprintf("%+v", IdentificarResponse))

	aEliminar := chains.NewLLMChain(llm, prompts.NewPromptTemplate(`You are tasked with identifying memory entries that should be deleted due to conflicts with newly retrieved facts.

Existing Memory:
{{.existing_memories}}

New Retrieved Facts Indicating Deletions:
{{.conflicts}}

Return a JSON object listing the IDs of memory entries to be deleted, along with an explanation for each deletion.

`, []string{
		"existing_memories",
		"conflicts",
	},
	))

	aEliminarResponse, err := chains.Call(ctx, aEliminar, map[string]interface{}{
		"existing_memories": serializedExistingMemories,
		"conflicts":         IdentificarResponse,
	})

	if err != nil {
		log.Panic(err)
	}

	c.debugPrint("aEliminarResponse: " + fmt.Sprintf("%+v", aEliminarResponse))

	elapsed := time.Since(st)
	c.debugPrint("Chain.MEMORY_UPDATER took: " + elapsed.String())

	return nil, fmt.Errorf("new error")

	// c.debugPrint("Using model: " + model)

	// messageHistory = updateMessageHistory(messageHistory, resp)

	// c.debugPrint("Output from LLM: " + fmt.Sprintf("%+v", messageHistory[1]))
	// Execute tool calls requested by the model
	// messageHistory = c.executeToolCalls(ctx, llm, messageHistory, resp)

	// return resp.Choices[0], nil
}
