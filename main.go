package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/matigumma/memGo/sqlitemanager"
	"github.com/matigumma/memGo/telemetry"
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
type LLM interface {
	GenerateResponse(messages []map[string]string, tools []Tool) (map[string]interface{}, error)
	GenerateResponseWithoutTools(messages []map[string]string) (string, error)
	// Add other methods as needed
}

// Embedder - Interface for Embedders (already defined, ensuring it's here for context)
type Embedder interface {
	Embed(text string) ([]float64, error)
	// Add other methods as needed
}

// --- Memory Class ---

// Memory - Corresponds to the Python Memory class
type Memory struct {
	config         MemoryConfig
	embeddingModel Embedder
	vectorStore    VectorStore
	telemetry      telemetry.AnonymousTelemetry
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

	phtelemetry, pherr := telemetry.NewAnonymousTelemetry("phc_eCRS68Q2koejazio0Umv93pwmGfwCH4uCa0dh1brRsI", "https://us.i.posthog.com", nil, nil)
	if pherr != nil {
		log.Fatalf("Error initializing telemetry: %v", pherr)
	}

	m := &Memory{
		config:         config,
		embeddingModel: embedder,
		vectorStore:    vectorStore,
		telemetry:      *phtelemetry,
		llm:            llm,
		db:             db,
		collectionName: config.VectorStore.Config["CollectionName"].(string),
	}
	m.telemetry.CaptureEvent("memGo.init", nil)
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

// Add creates a new memory
func (m *Memory) Add(data string, userID *string, agentID *string, runID *string, metadata map[string]interface{}, filters map[string]interface{}, prompt *string) (map[string]interface{}, error) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	embeddings, err := m.embeddingModel.Embed(data)
	if err != nil {
		return nil, fmt.Errorf("error embedding data: %w", err)
	}

	if filters == nil {
		filters = make(map[string]interface{})
	}
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

	currentPrompt := MEMORY_DEDUCTION_PROMPT
	if prompt != nil {
		currentPrompt = *prompt
	}
	formattedPrompt := fmt.Sprintf(currentPrompt, data, metadata)

	extractedMemories, err := m.llm.GenerateResponseWithoutTools([]map[string]string{
		{"role": "system", "content": "You are an expert at deducing facts, preferences and memories from unstructured text."},
		{"role": "user", "content": formattedPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("error generating response for memory deduction: %w", err)
	}
	log.Printf("Extracted memories: %s", extractedMemories)

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

	messages := getUpdateMemoryMessages(serializedExistingMemories, extractedMemories)

	tools := []Tool{ADD_MEMORY_TOOL, UPDATE_MEMORY_TOOL, DELETE_MEMORY_TOOL}
	response, err := m.llm.GenerateResponse(messages, tools)
	if err != nil {
		return nil, fmt.Errorf("error generating response with tools: %w", err)
	}

	toolCalls, ok := response["tool_calls"].([]interface{})
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
			m.telemetry.CaptureEvent("memGo.add.function_call", map[string]interface{}{"memory_id": functionResultID, "function_name": functionName})
		}
	}
	m.telemetry.CaptureEvent("memGo.add", nil)
	return map[string]interface{}{"message": "ok", "details": functionResults}, nil
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
	m.telemetry.CaptureEvent("memGo.get", map[string]interface{}{"memory_id": memoryID})
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

	m.telemetry.CaptureEvent("memGo.get_all", map[string]interface{}{"filters": len(filters), "limit": limit})
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

	m.telemetry.CaptureEvent("memGo.search", map[string]interface{}{"filters": len(filters), "limit": limit})
	embeddings, err := m.embeddingModel.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %w", err)
	}

	memories, err := m.vectorStore.Search(embeddings, limit, filters)
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
	m.telemetry.CaptureEvent("memGo.update", map[string]interface{}{"memory_id": memoryID})
	_, err := m.updateMemoryTool(memoryID, data)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"message": "Memory updated successfully!"}, nil
}

// Delete deletes a memory by ID
func (m *Memory) Delete(memoryID string) (map[string]interface{}, error) {
	m.telemetry.CaptureEvent("memGo.delete", map[string]interface{}{"memory_id": memoryID})
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

	m.telemetry.CaptureEvent("memGo.delete_all", map[string]interface{}{"filters": len(filters)})
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
	m.telemetry.CaptureEvent("memGo.history", map[string]interface{}{"memory_id": memoryID})
	return m.db.GetHistory(memoryID)
}

func (m *Memory) createMemoryTool(args map[string]interface{}) (string, error) {
	data, ok := args["data"].(string)
	if !ok {
		return "", errors.New("data not found or not a string")
	}
	metadata, ok := args["metadata"].(map[string]interface{})
	if !ok {
		metadata = make(map[string]interface{})
	}
	log.Printf("Creating memory with data=%s", data)

	embeddings, err := m.embeddingModel.Embed(data)
	if err != nil {
		return "", fmt.Errorf("error embedding data: %w", err)
	}

	memoryID := uuid.New().String()
	metadata["data"] = data

	hasher := md5.New()
	hasher.Write([]byte(data))
	metadata["hash"] = hex.EncodeToString(hasher.Sum(nil))

	pacific, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", fmt.Errorf("error loading timezone: %w", err)
	}
	metadata["created_at"] = time.Now().In(pacific).Format(time.RFC3339)

	err = m.vectorStore.Insert([][]float64{embeddings}, []string{memoryID}, []map[string]interface{}{metadata})
	if err != nil {
		return "", fmt.Errorf("error inserting into vector store: %w", err)
	}

	createdAt, ok := metadata["created_at"].(string)
	if !ok {
		return "", errors.New("created_at not found or not a string")
	}

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

	embeddings, err := m.embeddingModel.Embed(data)
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
	m.telemetry.CaptureEvent("memGo.reset", nil)
	return nil
}

// Chat - Placeholder
func (m *Memory) Chat(query string) error {
	return errors.New("chat function not implemented yet")
}
