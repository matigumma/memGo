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
	"github.com/matigumma/memGo/tools"
	"github.com/matigumma/memGo/utils"
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

func (c *Chain) PATTERNS_ATTENTION(data string) (map[string]interface{}, error) {
	c.debugPrint("Chain.PATTERNS_ATTENTION")

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

	prompt := `Eres un asistente experto en análisis de patrones en conversaciones y contextos. Tu objetivo es analizar inputs siguiendo estos pasos:

1. **Identificación de Patrones Lingüísticos:**
   - Evalúa el tono, estructura y estilo del mensaje (informativo, persuasivo, directo, indirecto, etc.).
   - Identifica frases o palabras clave que indiquen intenciones específicas.

2. **Reconocimiento de Patrones Contextuales:**
   - Analiza el entorno o situación implícita en el mensaje (contexto profesional, técnico, social, etc.).
   - Considera factores mencionados o subyacentes como dudas, limitaciones, oportunidades o condiciones específicas.

3. **Reconocimiento Basado en Intención:**
   - Determina las intenciones explícitas del mensaje.
   - Identifica posibles motivaciones implícitas o subtextos detrás de las palabras.
   - Considera si el mensaje busca informar, persuadir, pedir ayuda, tomar acción o validar algo.

4. **Patrones de Respuesta y Acción:**
   - Define qué tipo de respuesta o acción podría esperarse según el contenido del mensaje.
   - Evalúa si hay solicitudes implícitas o explícitas para quienes reciben el mensaje.

5. **Clasificación Final:**
   - Resume el propósito del mensaje en términos de intenciones principales y secundarias.
   - Identifica áreas clave de acción o sugerencias relevantes para responder al input.

Provee siempre un análisis claro y detallado usando este marco, asegurándote de estructurar el análisis con subtítulos para cada paso. Responde de forma profesional y objetiva.
Responde en formato JSON
`
	messages := []llms.MessageContent{}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, prompt))
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, data))

	ctx := context.Background()

	/* ====== GENERATE CONTENT ====== */
	out, err := llm.GenerateContent(ctx, messages, llms.WithJSONMode())
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	/* ====== DEBUG ====== */
	c.debugPrint("Using model: " + model)
	c.debugPrint("Output from LLM: " + fmt.Sprintf("%+v", out.Choices[0].Content))

	genInfo := out.Choices[0].GenerationInfo
	promptTokens, ok1 := genInfo["PromptTokens"].(int)
	completionTokens, ok2 := genInfo["CompletionTokens"].(int)
	if !ok1 || !ok2 {
		log.Printf("PromptTokens or CompletionTokens not found in GenerationInfo: %+v", genInfo)
	}

	totalTokens := int(promptTokens + completionTokens)
	c.debugPrint("Total Tokens: " + strconv.Itoa(totalTokens))
	c.debugPrint("Token Cost: " + fmt.Sprintf("%.6f", utils.EstimateCost(model, int(promptTokens), int(completionTokens))))

	/* ====== OUTPUT FORMAT ====== */
	parsedOutput := out.Choices[0].Content

	var result map[string]interface{}
	// stores the parsedOutput in the result value pointer
	err = json.Unmarshal([]byte(parsedOutput), &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	elapsed := time.Since(st)
	c.debugPrint("Chain.PATTERNS_ATTENTION took: " + elapsed.String())

	return result, nil
}

func (c *Chain) MEMORY_DEDUCTION(data string) (map[string]interface{}, error) {
	c.debugPrint("Chain.MEMORY_DEDUCTION")
	st := time.Now()

	/* ====== SETTINGS ====== */
	ctx := context.Background()

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

	strPrompt, err := prompt.Format(map[string]any{
		"conversation": data,
	})
	if err != nil {
		return nil, fmt.Errorf("error formatting prompt: %w", err)
	}

	messages := []llms.MessageContent{}
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, strPrompt))

	/* ====== CHAIN ====== */
	// llmChain := chains.NewLLMChain(llm, prompt)
	// out, err := chains.Call(ctx, llmChain, map[string]any{
	// 	"conversation": messages,
	// })

	/* ====== GENERATE CONTENT ====== */
	out, err := llm.GenerateContent(ctx, messages, llms.WithJSONMode())
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	/* ====== DEBUG ====== */
	c.debugPrint("Using model: " + model)
	c.debugPrint("Output from LLM: " + fmt.Sprintf("%+v", out.Choices[0].Content))

	genInfo := out.Choices[0].GenerationInfo
	promptTokens, ok1 := genInfo["PromptTokens"].(int)
	completionTokens, ok2 := genInfo["CompletionTokens"].(int)
	if !ok1 || !ok2 {
		log.Printf("PromptTokens or CompletionTokens not found in GenerationInfo: %+v", genInfo)
	}

	totalTokens := int(promptTokens + completionTokens)
	c.debugPrint("Total Tokens: " + strconv.Itoa(totalTokens))
	c.debugPrint("Token Cost: " + fmt.Sprintf("%.6f", utils.EstimateCost(model, int(promptTokens), int(completionTokens))))

	/* ====== OUTPUT FORMAT ====== */
	// parsedOutput, ok := out["text"].(string)
	parsedOutput := out.Choices[0].Content
	// .(string)
	// if !ok {
	// 	return nil, fmt.Errorf("failed to parse output text")
	// }

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

	prompt := prompts.NewPromptTemplate(
		p.MEMORY_UPDATER_FOR_EXISTING_AND_RELEVANT,
		[]string{
			"existing_memories",
			"relevantFactsText",
		},
	)

	/* ====== DATA FORMAT ====== */

	serializedExistingMemories := make([]map[string]interface{}, len(existingMemories))
	for i, item := range existingMemories {
		if item.Score != nil {
			serializedItem := map[string]interface{}{
				// "real_id": item.ID,
				"memory_id": item.ID,
				"memory":    item.Memory,
				"score":     *item.Score,
			}
			serializedExistingMemories[i] = serializedItem
			c.debugPrint("SRZd memory: " + fmt.Sprintf("%+v", serializedItem))
		}
	}

	// var newFacts []string
	var relevantFactsText string
	for i, fact := range relevantFacts {
		if factStr, ok := fact.(string); ok {
			relevantFactsText += fmt.Sprintf("%d - %s\n ", i+1, factStr)
			c.debugPrint("RLVNt fact: " + factStr)
		} else {
			c.debugPrint(fmt.Sprintf("Skipping non-string fact: %v", fact))
		}
	}

	strPrompt, err := prompt.Format(map[string]any{
		"existing_memories": serializedExistingMemories,
		"relevantFactsText": relevantFactsText,
	})
	if err != nil {
		return nil, fmt.Errorf("error formatting prompt: %w", err)
	}

	messages := []llms.MessageContent{}
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, "use available tools depending on status for NEW use `add_memory` tool and as data argument use 'fact'. for EXTEND use `update_memory` and as data argument use 'updated_memory'. for CONFLICT use `update_memory` tool and as data argument use 'updated_memory' and request at the same time a `delete_memory` tool for those invalidated memories."))
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, strPrompt))

	/* ====== GENERATE CONTENT ====== */

	out, err := llm.GenerateContent(ctx, messages, llms.WithTools([]models.Tool{tools.ADD_MEMORY_TOOL, tools.UPDATE_MEMORY_TOOL, tools.DELETE_MEMORY_TOOL}))
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	/* ====== OUTPUT SCHEMA ====== */

	/*
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

	/* ====== DEBUG ====== */
	c.debugPrint("Using model: " + model)
	c.debugPrint("Output from LLM: " + fmt.Sprintf("%+v", out.Choices[0].Content))

	genInfo := out.Choices[0].GenerationInfo
	promptTokens, ok1 := genInfo["PromptTokens"].(int)
	completionTokens, ok2 := genInfo["CompletionTokens"].(int)
	if !ok1 || !ok2 {
		log.Printf("PromptTokens or CompletionTokens not found in GenerationInfo: %+v", genInfo)
	}

	totalTokens := int(promptTokens + completionTokens)
	c.debugPrint("Total Tokens: " + strconv.Itoa(totalTokens))
	c.debugPrint("Token Cost: " + fmt.Sprintf("%.6f", utils.EstimateCost(model, int(promptTokens), int(completionTokens))))

	/*
		// Iterate through the RelevanciaResponse to filter facts
		// newrelevantFactsText := ""
		newFacts = make([]string, 0)
		// textResponse, ok := RelevanciaResponse["text"].(string)
		// if !ok {
		// 	log.Panic("RelevanciaResponse['text'] is not a map")
		// }
		textResponse := out.Choices[0].Content

		textResponse = strings.Trim(textResponse, "```json")
		textResponse = strings.Trim(textResponse, "```")

		var textResponseResult map[string]interface{}
		err = json.Unmarshal([]byte(textResponse), &textResponseResult)
		if err != nil {
			log.Panic("Error parsing JSON: ", err)
		}

		idsToDeleteFromResponseResult := make([]string, 0)

		result, ok := textResponseResult["results"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no 'results' in MEMORY_UPDATER_FOR_EXISTING_AND_RELEVANT")
		}

		for key, value := range result {

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
						if _, ok := fact.(string); ok {
							// newrelevantFactsText += fmt.Sprintf("%s - %s\n ", key, factStr)
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
							result[key] = map[string]interface{}{
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
							result[key] = map[string]interface{}{
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
			delete(result, id)
		}
	*/

	// c.debugPrint("NEW FACTS: " + strings.Join(newFacts, "\n"))
	// c.debugPrint("NEW RELEVANT FACTS: " + newrelevantFactsText)
	// relevantFactsText = newrelevantFactsText // lo piso...

	// c.debugPrint("FINAL RelevanciaResponse: " + fmt.Sprintf("%+v", result))

	/*
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
	*/

	elapsed := time.Since(st)
	c.debugPrint("Chain.MEMORY_UPDATER took: " + elapsed.String())

	// c.debugPrint("Using model: " + model)

	// messageHistory = updateMessageHistory(messageHistory, resp)

	// c.debugPrint("Output from LLM: " + fmt.Sprintf("%+v", messageHistory[1]))
	// Execute tool calls requested by the model
	// messageHistory = c.executeToolCalls(ctx, llm, messageHistory, resp)

	return out.Choices[0], nil
}
