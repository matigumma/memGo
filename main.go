package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/matigumma/memGo/chains"
	"github.com/matigumma/memGo/sqlitemanager"
	"github.com/matigumma/memGo/telemetry"
	"github.com/tmc/langchaingo/llms"
)

// --- tools

// --- prompts

// --- Configuration ---

// --- MemoryConfig ---

// --- Telemetry

// --- EmbedderConfig ---

// --- LlmConfig ---

// --- VectorStoreConfig ---

// --- MemoryItem ---

// --- Base Configurations ---

// SearchResult - Structure to hold search results from VectorStore
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// LLM - Interface for LLMs (already defined, ensuring it's here for context)
// type LLM interface{}

type LLM interface {
	GenerateResponse(messages []llms.MessageContent, tools []Tool, jsonMode bool, toolChoice string) (interface{}, error)
	// GenerateResponseWithoutTools(messages []map[string]string) (string, error)
	// Add other methods as needed
}

// Embedder - Interface for Embedders (already defined, ensuring it's here for context)
type Embedder interface {
	Embed(text string) ([]float64, []float32, error)
	// Add other methods as needed
}

// --- Memory Class ---

// Memory - Corresponds to the Python Memory class
type Memory struct {
	config         MemoryConfig
	embeddingModel Embedder
	vectorStore    VectorStore
	telemetry      *telemetry.AnonymousTelemetry
	llm            LLM
	db             *sqlitemanager.SQLiteManager
	collectionName string
}

// NewMemory creates a new Memory instance
func NewMemory(config MemoryConfig) *Memory {
	embedder, err := EmbedderFactory{}.Create(config.Embedder.Provider, config.Embedder.Config)
	if err != nil {
		log.Fatalf("Error creating embedder: %v", err)
	}
	vectorStore, err := VectorStoreFactory{}.Create(config.VectorStore.Provider, config.VectorStore.Config)
	if err != nil {
		log.Fatalf("Error creating vector store: %v", err)
	}
	llm, err := LlmFactory{}.Create(config.Llm.Provider, config.Llm.Config)
	if err != nil {
		log.Fatalf("Error creating LLM: %v", err)
	}
	db, err := sqlitemanager.NewSQLiteManager(config.HistoryDBPath)
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}

	// phtelemetry, pherr := telemetry.NewAnonymousTelemetry("phc_eCRS68Q2koejazio0Umv93pwmGfwCH4uCa0dh1brRsI", "https://us.i.posthog.com", nil, nil)
	// if pherr != nil {
	// 	log.Fatalf("Error initializing telemetry: %v", pherr)
	// }

	m := &Memory{
		//customPrompt: string //self.config.custom_prompt ?
		config:         config,      //MemoryConfig
		embeddingModel: embedder,    //EmbedderFactory
		vectorStore:    vectorStore, //VectorStoreFactory
		llm:            llm,         //LlmFactory
		db:             db,          //SQLiteManager
		telemetry:      nil,         //*phtelemetry,
		collectionName: "",
		// collectionName: config.VectorStore.Config["CollectionName"],
	}

	//v1.1 MemoryGraph?

	// m.telemetry.CaptureEvent("memGo.init", nil)
	return m
}

// FromConfig creates a Memory instance from a configuration map
func FromConfig(configMap map[string]interface{}) (*Memory, error) {
	configBytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %w", err)
	}

	var config MemoryConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Printf("Configuration validation error: %v", err)
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}
	return NewMemory(config), nil
}

// Add creates a new memory.
// the system extracts relevant facts and preferences and stores it across data stores:
// a vector database, a key-value database, and a graph database
func (m *Memory) Add(
	data string, // Messages to store in the memory. TODO: todo ver de manejar como un array de mensajes tambien
	userID *string, // ID of the user creating the memory. Defaults nil.
	agentID *string, // ID of the agent creating the memory. Defaults nil.
	runID *string, // ID of the run creating the memory. Defaults nil.
	metadata map[string]interface{}, // Metadata to store with the memory. Defaults nil
	filters map[string]interface{}, // Filters to apply to the search. Defaults nil
	prompt *string, // Prompt to use for memory deduction. Defaults nil.
) (map[string]interface{}, error) {
	fmt.Println("Memory.Add")
	// creo mapa de metadatos si no se pasa por parametro
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// creo mapa de filtros si no se pasa por parametro
	if filters == nil {
		filters = make(map[string]interface{})
	}

	// agrego filtros y metadatos
	if userID != nil {
		filters["user_id"] = *userID
		metadata["user_id"] = *userID
	}
	if agentID != nil {
		filters["agent_id"] = *agentID
		metadata["agent_id"] = *agentID
	}
	if runID != nil {
		filters["run_id"] = *runID
		metadata["run_id"] = *runID
	}

	// 1. check if at least one of userID, agentID, or runID is present
	if userID == nil && agentID == nil && runID == nil {
		return nil, errors.New("error: missing parameters, at least one of userID, agentID, or runID is required")
	}

	// en este paso prepara la ejecucion asincrona de la deduccion de la memoria en el vectorstore

	/* =============chain.MEMORY_DEDUCTION process============== */
	// Este paso obtiene informacion generalizada relevante de la data y devuelve in JSON bien estructurado.
	/* reduction test */

	fmt.Println("Raw Data: " + data)

	deductionAgent := chains.NewChain(true)
	Messages := []llms.MessageContent{}
	Messages = append(Messages, llms.TextParts(llms.ChatMessageTypeHuman, data))

	// 2. generates a prompt using the input messages and sends it to a Large Language Model (LLM) to retrieve new facts
	// _, err := reductionAgent.MEMORY_REDUCTION(Messages)
	deductios, err := deductionAgent.MEMORY_DEDUCTION(Messages)
	if err != nil {
		return nil, fmt.Errorf("error generating response for memory deduction: %w", err)
	}

	cantFacts := len(deductios["relevant_facts"].([]interface{}))
	// Check if there are any relevant facts to store
	if cantFacts == 0 {
		return map[string]interface{}{
			"message": "No memory added",
			"details": "no relevant facts found",
		}, nil
	}

	fmt.Println(fmt.Println("Relevant facts deducidos: " + strconv.Itoa(cantFacts)))
	/* end reduction test */

	// HASTA ACA TENEMOS:
	// -----------------
	// DATA string
	// MESSAGES []llms.MessageContent (de la data como ChatMessageTypeHuman)
	// DEDUCTIOS map[string]interface{} (relevants_facts, metadata)

	// DEDUCTIOS SCHEMA:
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
					"associations": {
						"type": "object",
						"properties": {
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
					},
					"sentiment": {
						"type": "string",
						"description": "The overall sentiment of the text, e.g., positive, negative, neutral"
					}
				}
			}
			}
		}
	*/

	/* ============= */

	/* Deduction test */
	/* end Deduction test */

	/* Search for related memories  */

	relevantFacts, ok := deductios["relevant_facts"].([]interface{})
	if !ok {
		return nil, errors.New("error: relevant_facts is not a list")
	}

	// el tamaño maximo es de la cantidad de relevant_facts * searchs limit de 5
	acumuladorMemoriasParaEvaluar := make([]MemoryItem, (len(relevantFacts) * 5))

	// metadata = deductios["metadata"]
	filterss := make(map[string]interface{})

	// Check if metadataMap is a valid map
	if metadataMap, ok := deductios["metadata"].(map[string]interface{}); ok {
		// Directly assign simple types
		if scope, ok := metadataMap["scope"].(string); ok {
			metadata["scope"] = scope
			// filterss["scope"] = scope
		}
		if sentiment, ok := metadataMap["sentiment"].(string); ok {
			metadata["sentiment"] = sentiment
			// filterss["sentiment"] = sentiment
		}

		// Convert related_entities to a slice of strings
		if relatedEntities, ok := metadataMap["related_entities"].([]interface{}); ok {
			var entities []string
			for _, entity := range relatedEntities {
				if strEntity, ok := entity.(string); ok {
					entities = append(entities, strEntity)
				}
			}
			metadata["related_entities"] = entities
			// filterss["related_entities"] = entities
		}

		// Convert related_events to a slice of strings
		if relatedEvents, ok := metadataMap["related_events"].([]interface{}); ok {
			var events []string
			for _, event := range relatedEvents {
				if strEvent, ok := event.(string); ok {
					events = append(events, strEvent)
				}
			}
			metadata["related_events"] = events
			// filterss["related_events"] = events
		}

		// Convert tags to a slice of strings
		if tags, ok := metadataMap["tags"].([]interface{}); ok {
			var tagList []string
			for _, tag := range tags {
				if strTag, ok := tag.(string); ok {
					tagList = append(tagList, strTag)
				}
			}
			metadata["tags"] = tagList
			// filterss["tags"] = tagList
		}
	}

	for fact_index, fact := range relevantFacts {
		factStr, ok := fact.(string)
		if !ok {
			log.Printf("Skipping non-string fact: %v", fact)
			continue
		}

		_, embeddings32, err := m.embeddingModel.Embed(factStr)
		if err != nil {
			log.Printf("Error embedding fact: %v", err)
			continue
		}

		existingMemoriesRaw, err := m.vectorStore.Search(embeddings32, 5, filterss)
		if err != nil {
			log.Printf("Error searching existing memories for fact: %v", err)
			continue
		}

		countExistingMemories := len(existingMemoriesRaw)

		if countExistingMemories == 0 {

			log.Printf("No existing memories found for fact index: %d", fact_index)
			// 2025/01/09 15:40:46 No existing memories found for fact index: 0
			// 2025/01/09 15:40:46 Creating memory with data=Está buscando material sobre ingeniería de prompts
			// result of creating memory tool: 1ca52a41-3393-4777-9c0a-2a25a039770e
			// en este caso deberian agregarse las nuevas memorias al vectorstore
			/*
				s, err := m.createMemoryTool(map[string]interface{}{"data": factStr, "metadata": metadata})
				if err != nil {
					return nil, fmt.Errorf("error creating memory tool: %w", err)
				}

				fmt.Println("result of creating memory tool: " + s)
			*/
			continue
		}

		fmt.Printf("Existing memories for fact %d: %d\n", fact_index, len(existingMemoriesRaw))
		fmt.Println(factStr)
		fmt.Println("---------------------------------------------")

		for i, mem := range existingMemoriesRaw {
			Score := &mem.Score
			hash := mem.Payload["hash"].(string)
			Metadata := mem.Payload
			Memory := mem.Payload["data"].(string)

			fmt.Printf("f:%d : m:%d - encontrado: %.6f, \n", fact_index, i, *Score)
			fmt.Printf("*Memory: %s\n", mem.Payload["data"])
			fmt.Printf("*Metadata: %s - %v - %v - %v\n", mem.Payload["scope"], mem.Payload["related_entities"], mem.Payload["related_events"], mem.Payload["tags"])
			fmt.Println("")
			// Existing memories for fact 0: 1
			// 0. Score: 1.000000, Metadata: map[agent_id:whatsapp created_at:2025-01-09T10:40:47-08:00 data:Está buscando material sobre ingeniería de prompts hash:6e53731be7ca1489e95b9e3cdcc3c58e user_id:Blas Briceño], Memory: Está buscando material sobre ingeniería de prompts
			// guardar en un acumulador para procesarlas luego
			acumuladorMemoriasParaEvaluar = append(acumuladorMemoriasParaEvaluar, MemoryItem{
				ID:       fmt.Sprintf("%d_%s", fact_index, mem.ID),
				Score:    Score,
				Memory:   Memory,
				Metadata: Metadata,
				Hash:     &hash,
			})
		}

		// INSERT_YOUR_CODE
		// To search in acumuladorMemoriasParaEvaluar by MemoryItem.ID, you can use a simple loop to iterate over the slice and check each item's ID.
		// Here's a function to perform the search:

		// Function to find a MemoryItem by ID
		// func findMemoryItemByID(id string, items []MemoryItem) *MemoryItem {
		// 	for _, item := range items {
		// 		if item.ID == id {
		// 			return &item
		// 		}
		// 	}
		// 	return nil
		// }

		// // Example usage
		// searchID := "some_id_to_search"
		// foundItem := findMemoryItemByID(searchID, acumuladorMemoriasParaEvaluar)
		// if foundItem != nil {
		// 	fmt.Printf("Found MemoryItem: %+v\n", *foundItem)
		// } else {
		// 	fmt.Println("MemoryItem not found")
		// }

		// _ = existingMemoriesRaw

		// Process existingMemoriesRaw as needed
		fmt.Println("")
	} // fin for range relevantFacts

	/* end Search for related memories test */

	// // 3. searches the vector store for existing memories similar to the new facts and retrieves their IDs and text
	// // create embeddings for the data
	// _, embeddings32, err := m.embeddingModel.Embed(data)
	// if err != nil {
	// 	return nil, fmt.Errorf("error embedding data: %w", err)
	// }

	// existingMemoriesRaw, err := m.vectorStore.Search(embeddings32, 5, filters)
	// if err != nil {
	// 	return nil, fmt.Errorf("error searching existing memories: %w", err)
	// }

	/*
		// deduccion de factos en el imput del usuario q valgan la pena
		currentPrompt := MEMORY_DEDUCTION_PROMPT // OJO ESTA VERSION ES SUPER REDUCIDA.. dejo en prompts el original
		if prompt != nil {
			// uso el prompt del usuario si lo pasa
			currentPrompt = *prompt
		}
		formattedPrompt := fmt.Sprintf(currentPrompt, data, metadata)

		inputMessages := []llms.MessageContent{}

		inputMessages = append(inputMessages, llms.TextParts(llms.ChatMessageTypeSystem, "You are an expert at deducing facts, preferences and memories from unstructured text."))
		inputMessages = append(inputMessages, llms.TextParts(llms.ChatMessageTypeHuman, formattedPrompt))

		// 2. generates a prompt using the input messages and sends it to a Large Language Model (LLM) to retrieve new facts
		//new_retrieved_facts
		extractedMemories, err := m.llm.GenerateResponse(inputMessages, nil, true)
		if err != nil {
			return nil, fmt.Errorf("error generating response for memory deduction: %w", err)
		}
	*/

	// log.Printf("Extracted factos: %+v \n", extractedMemories)
	/* ============END chain.MEMORY_DEDUCTION agent=============== */
	/*
		var newRetrievedFacts []interface{}
		extractedMemoriesStr, ok := extractedMemories.(string)
		if !ok {
			log.Printf("Error: extractedMemories is not a string")
			newRetrievedFacts = []interface{}{}
		} else {
			err = json.Unmarshal([]byte(extractedMemoriesStr), &newRetrievedFacts)
			if err != nil {
				log.Printf("Error in newRetrievedFacts: %v", err)
				newRetrievedFacts = []interface{}{}
			}
		}

		log.Printf("RetrievedFacts json factos: %+v \n", newRetrievedFacts)

		// 3. searches the vector store for existing memories similar to the new facts and retrieves their IDs and text
		// create embeddings for the data
		embeddings, err := m.embeddingModel.Embed(data)
		if err != nil {
			return nil, fmt.Errorf("error embedding data: %w", err)
		}

		existingMemoriesRaw, err := m.vectorStore.Search(embeddings, 5, filters)
		if err != nil {
			return nil, fmt.Errorf("error searching existing memories: %w", err)
		}

		existingMemories := make([]MemoryItem, len(existingMemoriesRaw))
		for i, mem := range existingMemoriesRaw {
			memoryItem := MemoryItem{
				ID:       mem.ID,
				Score:    &mem.Score,
				Metadata: mem.Payload,
				Memory:   mem.Payload["data"].(string), // Assuming "data" is always a string
			}
			existingMemories[i] = memoryItem
		}

		serializedExistingMemories := make([]map[string]interface{}, len(existingMemories))
		for i, item := range existingMemories {
			serializedItem := map[string]interface{}{
				"id":     item.ID,
				"memory": item.Memory,
				"score":  item.Score,
			}
			serializedExistingMemories[i] = serializedItem
		}
		log.Printf("Total existing memories: %d", len(existingMemories))

		messages := getUpdateMemoryMessages(serializedExistingMemories, newRetrievedFacts)

		tools := []Tool{ADD_MEMORY_TOOL, UPDATE_MEMORY_TOOL, DELETE_MEMORY_TOOL}

		// 4. generates another prompt to update the memories based on the new facts and sends it to the LLM
		response, err := m.llm.GenerateResponse([]llms.MessageContent{messages}, tools, false)
		if err != nil {
			return nil, fmt.Errorf("error generating response with tools: %w", err)
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error: response is not a map[string]interface{}")
		}

		// 5. processes the LLM's response, which contains actions to add, update, or delete memories
		toolCalls, ok := responseMap["tool_calls"].([]interface{})
		functionResults := make([]map[string]interface{}, 0)

		if ok {
			availableFunctions := map[string]func(map[string]interface{}) (string, error){
				"add_memory":    m.createMemoryTool,
				"update_memory": m.updateMemoryToolWrapper,
				"delete_memory": m.deleteMemoryTool,
			}

			for _, toolCallRaw := range toolCalls {
				toolCall, ok := toolCallRaw.(map[string]interface{})
				if !ok {
					log.Printf("Warning: Unexpected tool_call format: %+v", toolCallRaw)
					continue
				}

				functionName, ok := toolCall["name"].(string)
				if !ok {
					log.Printf("Warning: Function name not found or not a string in tool_call: %+v", toolCall)
					continue
				}

				functionToCall, ok := availableFunctions[functionName]
				if !ok {
					log.Printf("Warning: Function %s not found in available functions", functionName)
					continue
				}

				argumentsRaw, ok := toolCall["arguments"].(string)
				if !ok {
					log.Printf("Warning: Arguments not found or not a string in tool_call: %+v", toolCall)
					continue
				}

				var functionArgs map[string]interface{}
				err = json.Unmarshal([]byte(argumentsRaw), &functionArgs)
				if err != nil {
					log.Printf("Error unmarshaling function arguments: %v", err)
					continue
				}

				log.Printf("[openai_func] func: %s, args: %+v", functionName, functionArgs)

				if functionName == "add_memory" || functionName == "update_memory" {
					functionArgs["metadata"] = metadata
				}

				// 6. performs the actions on the memories, creating new ones, updating existing ones, or deleting them.
				functionResultID, err := functionToCall(functionArgs)
				if err != nil {
					log.Printf("Error calling function %s: %v", functionName, err)
					continue
				}

				functionResults = append(functionResults, map[string]interface{}{
					"id":    functionResultID,
					"event": trimMemorySuffix(functionName),
					"data":  functionArgs["data"],
				})
				// m.telemetry.CaptureEvent("memGo.add.function_call", map[string]interface{}{"memory_id": functionResultID, "function_name": functionName})
			}
		}
		// 7. returns a list of memories with their IDs, text, and events (ADD, UPDATE, DELETE, or NONE)
	*/
	// m.telemetry.CaptureEvent("memGo.add", nil)
	return map[string]interface{}{"message": "ok", "details": "functionResults"}, nil
}

func (m *Memory) updateMemoryToolWrapper(args map[string]interface{}) (string, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok {
		return "", errors.New("memory_id not found or not a string")
	}
	data, ok := args["data"].(string)
	if !ok {
		return "", errors.New("data not found or not a string")
	}
	return m.updateMemoryTool(memoryID, data)
}

// Get retrieves a memory by ID
func (m *Memory) Get(memoryID string) (map[string]interface{}, error) {
	// m.telemetry.CaptureEvent("memGo.get", map[string]interface{}{"memory_id": memoryID})
	memory, err := m.vectorStore.Get(memoryID)
	if err != nil {
		return nil, fmt.Errorf("error getting memory from vector store: %w", err)
	}
	if memory == nil {
		return nil, nil
	}

	filters := make(map[string]interface{})
	for _, key := range []string{"user_id", "agent_id", "run_id"} {
		if val, ok := memory.Payload[key]; ok {
			filters[key] = val
		}
	}

	memoryItem := map[string]interface{}{
		"id":         memory.ID,
		"memory":     memory.Payload["data"],
		"hash":       memory.Payload["hash"],
		"created_at": memory.Payload["created_at"],
		"updated_at": memory.Payload["updated_at"],
	}

	additionalMetadata := make(map[string]interface{})
	excludedKeys := map[string]bool{"user_id": true, "agent_id": true, "run_id": true, "hash": true, "data": true, "created_at": true, "updated_at": true}
	for k, v := range memory.Payload {
		if !excludedKeys[k] {
			additionalMetadata[k] = v
		}
	}
	if len(additionalMetadata) > 0 {
		memoryItem["metadata"] = additionalMetadata
	}

	result := mergeMaps(memoryItem, filters)
	return result, nil
}

// GetAll lists all memories
func (m *Memory) GetAll(userID *string, agentID *string, runID *string, limit int) ([]map[string]interface{}, error) {
	filters := make(map[string]interface{})
	if userID != nil {
		filters["user_id"] = *userID
	}
	if agentID != nil {
		filters["agent_id"] = *agentID
	}
	if runID != nil {
		filters["run_id"] = *runID
	}

	// m.telemetry.CaptureEvent("memGo.get_all", map[string]interface{}{"filters": len(filters), "limit": limit})
	memoriesList, err := m.vectorStore.List(filters, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing memories: %w", err)
	}

	var allMemories []map[string]interface{}
	excludedKeys := map[string]bool{"user_id": true, "agent_id": true, "run_id": true, "hash": true, "data": true, "created_at": true, "updated_at": true}

	for _, memories := range memoriesList {
		for _, mem := range memories {
			memoryItem := map[string]interface{}{
				"id":         mem.ID,
				"memory":     mem.Payload["data"],
				"hash":       mem.Payload["hash"],
				"created_at": mem.Payload["created_at"],
				"updated_at": mem.Payload["updated_at"],
			}
			for _, key := range []string{"user_id", "agent_id", "run_id"} {
				if val, ok := mem.Payload[key]; ok {
					memoryItem[key] = val
				}
			}
			additionalMetadata := make(map[string]interface{})
			for k, v := range mem.Payload {
				if !excludedKeys[k] {
					additionalMetadata[k] = v
				}
			}
			if len(additionalMetadata) > 0 {
				memoryItem["metadata"] = additionalMetadata
			}
			allMemories = append(allMemories, memoryItem)
		}
	}
	return allMemories, nil
}

// Search searches for memories
func (m *Memory) Search(query string, userID *string, agentID *string, runID *string, limit int, filters map[string]interface{}) ([]map[string]interface{}, error) {
	if filters == nil {
		filters = make(map[string]interface{})
	}
	if userID != nil {
		filters["user_id"] = *userID
	}
	if agentID != nil {
		filters["agent_id"] = *agentID
	}
	if runID != nil {
		filters["run_id"] = *runID
	}

	// m.telemetry.CaptureEvent("memGo.search", map[string]interface{}{"filters": len(filters), "limit": limit})
	_, embeddings32, err := m.embeddingModel.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %w", err)
	}

	memories, err := m.vectorStore.Search(embeddings32, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("error searching vector store: %w", err)
	}

	var searchResults []map[string]interface{}
	excludedKeys := map[string]bool{"user_id": true, "agent_id": true, "run_id": true, "hash": true, "data": true, "created_at": true, "updated_at": true}

	for _, mem := range memories {
		memoryItem := map[string]interface{}{
			"id":         mem.ID,
			"memory":     mem.Payload["data"],
			"hash":       mem.Payload["hash"],
			"created_at": mem.Payload["created_at"],
			"updated_at": mem.Payload["updated_at"],
			"score":      mem.Score,
		}
		for _, key := range []string{"user_id", "agent_id", "run_id"} {
			if val, ok := mem.Payload[key]; ok {
				memoryItem[key] = val
			}
		}
		additionalMetadata := make(map[string]interface{})
		for k, v := range mem.Payload {
			if !excludedKeys[k] {
				additionalMetadata[k] = v
			}
		}
		if len(additionalMetadata) > 0 {
			memoryItem["metadata"] = additionalMetadata
		}
		searchResults = append(searchResults, memoryItem)
	}
	return searchResults, nil
}

// Update updates a memory by ID
func (m *Memory) Update(memoryID string, data string) (map[string]interface{}, error) {
	// m.telemetry.CaptureEvent("memGo.update", map[string]interface{}{"memory_id": memoryID})
	_, err := m.updateMemoryTool(memoryID, data)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"message": "Memory updated successfully!"}, nil
}

// Delete deletes a memory by ID
func (m *Memory) Delete(memoryID string) (map[string]interface{}, error) {
	// m.telemetry.CaptureEvent("memGo.delete", map[string]interface{}{"memory_id": memoryID})
	_, err := m.deleteMemoryTool(map[string]interface{}{"memory_id": memoryID})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"message": "Memory deleted successfully!"}, nil
}

// DeleteAll deletes all memories based on filters
func (m *Memory) DeleteAll(userID *string, agentID *string, runID *string) (map[string]interface{}, error) {
	filters := make(map[string]interface{})
	if userID != nil {
		filters["user_id"] = *userID
	}
	if agentID != nil {
		filters["agent_id"] = *agentID
	}
	if runID != nil {
		filters["run_id"] = *runID
	}

	if len(filters) == 0 {
		return nil, errors.New("at least one filter is required to delete all memories. If you want to delete all memories, use the `Reset()` method")
	}

	// m.telemetry.CaptureEvent("memGo.delete_all", map[string]interface{}{"filters": len(filters)})
	memoriesList, err := m.vectorStore.List(filters, -1) // Get all matching memories
	if err != nil {
		return nil, fmt.Errorf("error listing memories for deletion: %w", err)
	}

	for _, memories := range memoriesList {
		for _, memory := range memories {
			_, err = m.deleteMemoryTool(map[string]interface{}{"memory_id": memory.ID})
			if err != nil {
				log.Printf("Error deleting memory %s: %v", memory.ID, err)
				// Consider whether to continue or return an error here
			}
		}
	}
	return map[string]interface{}{"message": "Memories deleted successfully!"}, nil
}

// History gets the history of changes for a memory by ID
func (m *Memory) History(memoryID string) ([]map[string]interface{}, error) {
	// m.telemetry.CaptureEvent("memGo.history", map[string]interface{}{"memory_id": memoryID})
	return m.db.GetHistory(memoryID)
}

func (m *Memory) createMemoryTool(args map[string]interface{}) (string, error) {
	// 1. extracts the data and metadata from the args map
	data, ok := args["data"].(string)
	if !ok {
		return "", errors.New("data not found or not a string")
	}
	metadata, ok := args["metadata"].(map[string]interface{})
	if !ok {
		metadata = make(map[string]interface{})
	}
	log.Printf("Creating memory with data=%s", data)

	// 2. embeds the data using the embeddingModel
	embeddings, _, err := m.embeddingModel.Embed(data)
	if err != nil {
		return "", fmt.Errorf("error embedding data: %w", err)
	}

	memoryID := uuid.New().String()
	metadata["data"] = data

	hasher := md5.New()
	hasher.Write([]byte(data))
	metadata["hash"] = hex.EncodeToString(hasher.Sum(nil))

	pacific, err := time.LoadLocation("America/Buenos_Aires")
	if err != nil {
		return "", fmt.Errorf("error loading timezone: %w", err)
	}
	metadata["created_at"] = time.Now().In(pacific).Format(time.RFC3339)

	// 3. inserts the embeddings, memoryID, and metadata into the vectorStore
	err = m.vectorStore.Insert([][]float64{embeddings}, []string{memoryID}, []map[string]interface{}{metadata})
	if err != nil {
		return "", fmt.Errorf("error inserting into vector store: %w", err)
	}

	createdAt, ok := metadata["created_at"].(string)
	if !ok {
		return "", errors.New("created_at not found or not a string")
	}

	// 4. adds a history entry to the db indicating that a memory with the given memoryID was added
	err = m.db.AddHistory(memoryID, nil, data, "ADD", &createdAt, nil, 0)
	if err != nil {
		log.Printf("Error adding history: %v", err) // Non-critical error
	}
	return memoryID, nil
}

func (m *Memory) updateMemoryTool(memoryID string, data string) (string, error) {
	log.Printf("Updating memory with memoryID=%s with data=%s", memoryID, data)

	existingMemory, err := m.vectorStore.Get(memoryID)
	if err != nil {
		return "", fmt.Errorf("error getting existing memory: %w", err)
	}
	if existingMemory == nil {
		return "", fmt.Errorf("memory with ID %s not found", memoryID)
	}
	prevValue := existingMemory.Payload["data"].(string)

	newMetadata := make(map[string]interface{})
	newMetadata["data"] = data
	newMetadata["hash"] = existingMemory.Payload["hash"]
	newMetadata["created_at"] = existingMemory.Payload["created_at"]

	pacific, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("error loading timezone: %w", err)
	}
	newMetadata["updated_at"] = time.Now().In(pacific).Format(time.RFC3339)

	for _, key := range []string{"user_id", "agent_id", "run_id"} {
		if val, ok := existingMemory.Payload[key]; ok {
			newMetadata[key] = val
		}
	}

	embeddings, _, err := m.embeddingModel.Embed(data)
	if err != nil {
		return "", fmt.Errorf("error embedding data: %w", err)
	}

	err = m.vectorStore.Update(memoryID, embeddings, newMetadata)
	if err != nil {
		return "", fmt.Errorf("error updating vector store: %w", err)
	}

	err = m.db.AddHistory(memoryID, &prevValue, data, "UPDATE", newMetadata["created_at"].(*string), newMetadata["updated_at"].(*string), 0)
	if err != nil {
		log.Printf("Error adding history: %v", err) // Non-critical error
	}
	return memoryID, nil
}

func (m *Memory) deleteMemoryTool(args map[string]interface{}) (string, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok {
		return "", errors.New("memory_id not found or not a string")
	}
	log.Printf("Deleting memory with memoryID=%s", memoryID)

	existingMemory, err := m.vectorStore.Get(memoryID)
	if err != nil {
		return "", fmt.Errorf("error getting existing memory for deletion: %w", err)
	}
	if existingMemory == nil {
		return "", fmt.Errorf("memory with ID %s not found for deletion", memoryID)
	}
	prevValue := existingMemory.Payload["data"].(string)

	err = m.vectorStore.Delete(memoryID)
	if err != nil {
		return "", fmt.Errorf("error deleting from vector store: %w", err)
	}

	pacific, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("error loading timezone: %w", err)
	}

	now := time.Now().In(pacific).Format(time.RFC3339)
	err = m.db.AddHistory(memoryID, &prevValue, "", "DELETE", &now, &now, 1)
	if err != nil {
		log.Printf("Error adding history: %v", err) // Non-critical error
	}
	return "", nil
}

// Reset resets the memory store
func (m *Memory) Reset() error {
	err := m.vectorStore.DeleteCol()
	if err != nil {
		return fmt.Errorf("error deleting vector store collection: %w", err)
	}
	err = m.db.Reset()
	if err != nil {
		return fmt.Errorf("error resetting database: %w", err)
	}
	// m.telemetry.CaptureEvent("memGo.reset", nil)
	return nil
}

// Chat - Placeholder
func (m *Memory) Chat(query string) error {
	return errors.New("chat function not implemented yet")
}

func main() {
	MemoryConfig := NewMemoryConfig()

	m := NewMemory(MemoryConfig)

	// userId := "pgp"
	// agentId := "carolina"
	// runId := "entrevista-1"

	/* ===mock data from transcription interview=== */
	/*
		// Read the content of transcription.txt
		transcriptionContent, err := os.ReadFile("/Users/agrosistemas/Mati/test-golang/transcription.txt")
		if err != nil {
			log.Panicf("error reading transcription file: %v", err)
		}

		// Convert the content to a string
		transcriptionText := string(transcriptionContent)
	*/
	/* ===mock data from transcription interview=== */

	/* ===chunked data from transcription audio files=== */
	/*
		// Read the content of chunk_transcription.json
		chunkFileContent, err := os.ReadFile("/Users/agrosistemas/Mati/test-golang/chunked_transcription.json")
		if err != nil {
			log.Panicf("error reading chunk transcription file: %v", err)
		}

		// Parse the JSON content
		var chunkData struct {
			Chunks []string `json:"chunks"`
		}
		err = json.Unmarshal(chunkFileContent, &chunkData)
		if err != nil {
			log.Panicf("error unmarshaling chunk transcription data: %v", err)
		}

		metadata := map[string]interface{}{
			"fecha": time.Now().Format(time.RFC3339),
		}

		// Iterate over each chunk and process with m.Add
		for index, chunk := range chunkData.Chunks {
			fmt.Println("chunk index: ", index)
			metadata["chunk_[120]_index"] = index
			_, err := m.Add(chunk, &userId, &agentId, &runId, metadata, nil, nil) // Pass metadata directly without using &
			if err != nil {
				log.Printf("Error adding memory for chunk: %v", err)
			}
		}
	*/
	/* ===chunked data from transcription audio files=== */

	userId := "Blas Briceño"
	agentId := "whatsapp"
	// runId := "entrevista-1"

	text := "Hola, me contactó un posible cliente que necesita implementar un chatboot que participando de un grupo de whatsapp analice las conversaciones para encontrar cierta información y después al encontrarse con ciertos parámetros contacte por whatsapp a números que se encuentran en la conversación misma y le mande un mensaje y tal vez le permita ingresar información que debe ser persistida en una base de datos. En ITR podemos hacer este desarrollo, pero no me cierra el tamaño del cliente / posibilidades económicas. Si a alguien le interesa contácteme por privado para ponerlo en contacto con el cliente"
	// text := "vengo acá a recordarles que mañana a las 17 hacemos el brainstorming y reunión de encuentro, con los que puedan sumarse."
	// text := "Buen día, consultita en el grupo ¿han socializado algún material sobre ingeniería de prompts?"
	// text += "Para darles contexto estoy preparando un documento de prompts para que le sirva a 3 equipos (copy,diseño y comtent) para la empresa en la que trabajo. De modo que quería tener otros recursos bibliográficas para ampliar el material"
	// text := "Gus, creo que podemos arrancar con una esa semana, y después a fin de enero la continuamos con una más.. no creo que con una sola reunión semejante profusión de ideas se pueda hacer converger de una"
	// text := "Hola, buen dia"
	// text := "hay que quedar un monto para 10 siguientes y te envió por crypto. El anterior fueron $100 equivalentes en crypto por 10 adicionales, lo repetimos?"
	res, err := m.Add(
		text,     // data
		&userId,  // user_id
		&agentId, // agent_id
		nil,      // run_id
		nil,      // metadata
		nil,      // filtros
		nil,      // custom prompt
	)
	if err != nil {
		log.Fatalf("Error adding memory: %v", err)
	}

	fmt.Printf("Memory add response: %+v\n", res)

	// search, err := m.Search("hello", &userId, nil, nil, 5, nil)
	// if err != nil {
	// 	log.Fatalf("Error searching memory: %v", err)
	// }
	// fmt.Printf("Search results: %+v\n", search)

}
