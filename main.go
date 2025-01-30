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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/matigumma/memGo/chains"
	"github.com/matigumma/memGo/models"
	"github.com/matigumma/memGo/sqlitemanager"
	"github.com/matigumma/memGo/telemetry"
	"github.com/matigumma/memGo/utils"
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

/*
class Record(BaseModel):
    """
    Point data
    """

    id: "ExtendedPointId" = Field(..., description="Point data")
    payload: Optional["Payload"] = Field(default=None, description="Payload - values assigned to the point")
    vector: Optional["VectorStruct"] = Field(default=None, description="Vector of the point")
    shard_key: Optional["ShardKey"] = Field(default=None, description="Shard Key")
    order_value: Optional["OrderValue"] = Field(default=None, description="Point data")
*/

// LLM - Interface for LLMs (already defined, ensuring it's here for context)
// type LLM interface{}

type LLM interface {
	GenerateResponse(messages []llms.MessageContent, tools []models.Tool, jsonMode bool, toolChoice string) (interface{}, error)
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
	debug          bool
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
		debug:          false,
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
	gc *gin.Context,
) (map[string]interface{}, error) {
	fmt.Println("Memory.Add")
	/* ====== VALICACIONES ====== */
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

	// 1. check if at least ONE of userID, agentID, or runID is present
	if userID == nil && agentID == nil && runID == nil {
		return nil, errors.New("error: missing parameters, at least one of userID, agentID, or runID is required")
	}

	// en este paso prepara la ejecucion asincrona de la deduccion de la memoria en el vectorstore

	utils.DebugPrint("Raw INPUT Data: "+data, m.debug, gc)

	/* ============= chain.MEMORY_DEDUCTION process ============== */

	// Este paso obtiene informacion generalizada relevante de la data de la memoria guardada en el VectorStore
	deductionChain := chains.NewChain(m.debug, gc)

	/*
		// PATTERNS_ATTENTION busca patrones en el mensaje y devuelve un json con las clasificaciones

		patterns, err := deductionChain.PATTERNS_ATTENTION(data)
		if err != nil {
			return nil, fmt.Errorf("error generating response for PATTERNS_ATTENTION: %w", err)
		}
	*/

	// 2. generates a prompt using the input data and

	// sends it to a Large Language Model (LLM) to retrieve new relevant facts
	deduction, err := deductionChain.MEMORY_DEDUCTION(data)
	if err != nil {
		return nil, fmt.Errorf("error generating response for MEMORY_DEDUCTION: %w", err)
	}
	/* ====== DEDUCTION OUTPUT ====== */

	/*
		{
			"relevant_facts": [
				"Recibió un contacto de un posible cliente interesado en implementar un chatbot.",
				"El chatbot debe analizar conversaciones de WhatsApp para extraer información y contactar números en las conversaciones.",
				"Considera que ITR puede desarrollar el proyecto, pero tiene dudas sobre el tamaño del cliente y sus posibilidades económicas.",
				"Está abierto a que otros se pongan en contacto para conectar con el cliente."
			],
			"metadata": {
				"scope": "negocios",
				"sentiment": "neutral",
				"related_entities": ["cliente", "chatbot", "WhatsApp", "ITR"],
				"related_events": ["contacto con cliente", "desarrollo de software"],
				"tags": ["negocios", "tecnología", "chatbot", "WhatsApp"]
			}
		}
	*/

	/* ====== VALIDATION DEDUCTION OUTPUT ====== */
	relevantFacts, ok := deduction["relevant_facts"].([]interface{})
	if !ok {
		// print this error in case for inspection (maybe should log it in a file...)
		fmt.Println(fmt.Errorf("error: relevant_facts is not a list after MEMORY_DEDUCTION\n %v", deduction))

		return map[string]interface{}{
			"message": "No memory added",
			"details": "no relevant facts found",
		}, nil
	}
	cantFacts := len(relevantFacts)
	if cantFacts == 0 {
		return map[string]interface{}{
			"message": "No memory added",
			"details": "no relevant facts found",
		}, nil
	}

	utils.DebugPrint(fmt.Sprintf("# RELEVANT FACTS DEDUCIDOS: %s\n", strconv.Itoa(cantFacts)), m.debug, gc)

	// el tamaño maximo es de la cantidad de relevant_facts * searchs limit de 5
	acumuladorMemoriasParaEvaluar := make([]models.MemoryItem, 0, len(relevantFacts)*5)

	// metadata = deduction["metadata"]
	filterss := filters

	// VALIDO METADATA Y AGREGO LA METADATA GENERADA DE LA DEDUCCION
	if metadataMap, ok := deduction["metadata"].(map[string]interface{}); ok {
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

	/* ====== SIMILARITY SEARCH FOR EVERY FACT OF DEDUCTIONS =====  */
	for fact_index, fact := range relevantFacts {
		factStr, ok := fact.(string)
		if !ok {
			utils.DebugPrint(fmt.Sprintf("Error skipping non-string fact: %v\n at index: %d", fact, fact_index), m.debug, gc)
			continue
		}

		_, embeddings32, err := m.embeddingModel.Embed(factStr)
		if err != nil {
			utils.DebugPrint(fmt.Sprintf("Error embedding fact: %v\n at index: %d \nerr: %v", fact, fact_index, err), m.debug, gc)
			return nil, fmt.Errorf("error embedding fact")
		}

		/* ====== SEARCH FOR max(5) EXISTING MEMORIES IN VS WITH Filters ===== */
		existingMemoriesRaw, err := m.vectorStore.Search(embeddings32, 5, filterss)
		if err != nil {
			utils.DebugPrint(fmt.Sprintf("Error searching existing memories for fact: %v\n at index: %d\nerr: %v", fact, fact_index, err), m.debug, gc)
			return nil, fmt.Errorf("error searching existing memories")
		}

		/* ====== SEARCH OUTPUT ====== */

		/*
			[
			    {
			        "ID": {
			            "uuid": "a039176a-3aae-43e1-ab55-e5cfda3c6777"
			        },
			        "Score": 0.9994954466819763,
			        "Payload": {
			            "agent_id": "whatsapp",
			            "created_at": "2025-01-10T11:32:53-03:00",
			            "data": "Un posible cliente necesita implementar un chatbot que analice conversaciones de WhatsApp.",
			            "hash": "4b2effcf93b2591e0d933f7681974cba",
			            "related_entities": [
			                "cliente",
			                "chatbot",
			                "WhatsApp"
			            ],
			            "related_events": [
			                "implementación de tecnología",
			                "análisis de conversaciones"
			            ],
			            "scope": "tecnología",
			            "sentiment": "neutral",
			            "tags": [
			                "desarrollo",
			                "tecnología",
			                "chatbot",
			                "WhatsApp"
			            ],
			            "user_id": "Blas Briceño"
			        }
			    },
			    {
			        "ID": {
			            "uuid": "0b54783e-8048-4f56-950b-b711ba40eb64"
			        },
			        "Score": 0.9184341430664062,
			        "Payload": {
			            "agent_id": "whatsapp",
			            "created_at": "2025-01-10T09:56:55-03:00",
			            "data": "Recibió un contacto de un posible cliente interesado en implementar un chatbot.",
			            "hash": "e54348e74901e683f8f9fbe3ce5f5e1e",
			            "related_entities": [
			                "cliente",
			                "chatbot",
			                "WhatsApp"
			            ],
			            "related_events": [
			                "contacto con cliente",
			                "desarrollo de software"
			            ],
			            "scope": "negocios",
			            "sentiment": "neutral",
			            "tags": [
			                "negocios",
			                "tecnología",
			                "chatbot",
			                "WhatsApp"
			            ],
			            "user_id": "Blas Briceño"
			        }
			    }
			]
		*/

		/* ====== VALIDATION SEARCH OUTPUT ====== */
		countExistingMemories := len(existingMemoriesRaw)
		if countExistingMemories == 0 {
			utils.DebugPrint(fmt.Sprintf("No existing memories found for fact index: %d", fact_index), m.debug, gc)
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
			utils.DebugPrint("\n", m.debug, gc)
			continue
		}

		utils.DebugPrint(fmt.Sprintf("Existing memories for fact %d: %d\n", fact_index, len(existingMemoriesRaw)), m.debug, gc)
		utils.DebugPrint(factStr, m.debug, gc)
		utils.DebugPrint("---------------------------------------------", m.debug, gc)

		// De aca en adelante encontre memorias en el vectorstore para con este facto.
		for i, mem := range existingMemoriesRaw {
			Score := &mem.Score
			hash := mem.Payload["hash"].(string)
			Metadata := mem.Payload
			Memory := mem.Payload["data"].(string)

			utils.DebugPrint(fmt.Sprintf("f:%d : m:%d - encontrado: %.6f, \n", fact_index, i, *Score), m.debug, gc)
			utils.DebugPrint(fmt.Sprintf("*Memory: %s\n", mem.Payload["data"]), m.debug, gc)
			utils.DebugPrint(fmt.Sprintf("*Metadata: %s - %v - %v - %v\n", mem.Payload["scope"], mem.Payload["related_entities"], mem.Payload["related_events"], mem.Payload["tags"]), m.debug, gc)
			utils.DebugPrint("\n", m.debug, gc)
			// Existing memories for fact 0: 1
			// 0. Score: 1.000000, Metadata: map[agent_id:whatsapp created_at:2025-01-09T10:40:47-08:00 data:Está buscando material sobre ingeniería de prompts hash:6e53731be7ca1489e95b9e3cdcc3c58e user_id:Blas Briceño], Memory: Está buscando material sobre ingeniería de prompts
			// guardar en un acumulador para procesarlas luego
			// creo un MemoryItem por cada una y las acumulo para evaluar luego.
			acumuladorMemoriasParaEvaluar = append(acumuladorMemoriasParaEvaluar, models.MemoryItem{
				ArrangeIndex: fact_index,
				ID:           mem.ID,
				Score:        Score,
				Memory:       Memory,
				Metadata:     Metadata,
				Hash:         &hash,
			})
		}

		utils.DebugPrint("\n", m.debug, gc)
	}

	utils.DebugPrint(fmt.Sprintf("# MEMORIAS ENCONTRADAS PARA EVALUAR: %s\n", strconv.Itoa(len(acumuladorMemoriasParaEvaluar))), m.debug, gc)

	/* ============ chain.MEMORY_UPDATER process =============== */

	actionsAgent := chains.NewChain(true, gc)

	// 2. generates a prompt using the input messages and sends it to
	// a Large Language Model (LLM) to retrieve new facts
	responseMap, err := actionsAgent.MEMORY_UPDATER(acumuladorMemoriasParaEvaluar, relevantFacts)
	if err != nil {
		return nil, fmt.Errorf("error generating response for MEMORY_UPDATER: %w", err)
	}

	// split memory updater into little steps

	utils.DebugPrint(fmt.Sprintln("PHASE 2: MEMORY_UPDATER OK"), m.debug, gc)

	// 5. processes the LLM's response, which contains actions to add, update, or delete memories
	toolCalls := responseMap.ToolCalls
	functionResults := make([]map[string]interface{}, 0)

	availableFunctions := map[string]func(map[string]interface{}) (string, error){
		"add_memory": m.createMemoryTool,
		"update_memory": func(args map[string]interface{}) (string, error) {
			return m.updateMemoryToolWrapper(args, gc)
		},
		"delete_memory": m.deleteMemoryTool,
		"no_op_memory":  func(m map[string]interface{}) (string, error) { return "", nil },
		"resolve_memory_conflict": func(args map[string]interface{}) (string, error) {
			m1, ok1 := args["memory1"].(map[string]interface{})
			m2, ok2 := args["memory2"].(map[string]interface{})
			strategy, ok3 := args["strategy"].(string)

			if !ok1 || !ok2 || !ok3 {
				return "", errors.New("invalid arguments")
			}

			utils.DebugPrint(fmt.Sprint("m1: ", m1), m.debug, gc)
			utils.DebugPrint(fmt.Sprint("m2: ", m2), m.debug, gc)
			utils.DebugPrint(fmt.Sprint("strategy: ", strategy), m.debug, gc)

			// Implement the conflict resolution logic here
			utils.DebugPrint("resolve_memory_conflict logic executed", m.debug, gc)

			return "resolved_memory_id", nil
		},
	}

	for _, toolCall := range toolCalls {
		functionName := toolCall.FunctionCall.Name
		utils.DebugPrint(fmt.Sprintf("Processing function: %s", functionName), m.debug, gc)
		functionToCall, ok := availableFunctions[functionName]
		if !ok {
			utils.DebugPrint(fmt.Sprintf("Warning: Function %s not found in available functions", functionName), m.debug, gc)
			continue
		}

		if functionName == "no_op_memory" {
			continue
		}

		argumentsRaw := toolCall.FunctionCall.Arguments

		var functionArgs map[string]interface{}
		err = json.Unmarshal([]byte(argumentsRaw), &functionArgs)
		if err != nil {
			log.Printf("Error unmarshaling function arguments: %v", err)
			continue
		}

		if functionName == "delete_memory" || functionName == "update_memory" {
			indexStr := functionArgs["memory_id"].(string)

			utils.DebugPrint(fmt.Sprintf("functionName: %s", functionName), m.debug, gc)
			utils.DebugPrint(fmt.Sprintf("indexStr: %s", indexStr), m.debug, gc)

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("error converting memory index to int: %w", err)
			}
			real_id := acumuladorMemoriasParaEvaluar[index].ID
			utils.DebugPrint(fmt.Sprintf("real_id: %s", real_id), m.debug, gc)

			functionArgs["memory_id"] = real_id
		}

		utils.DebugPrint(fmt.Sprintf("[openai_func] func: %s\nargs: %+v\n", functionName, functionArgs), m.debug, gc)

		if functionName == "add_memory" || functionName == "update_memory" {
			functionArgs["metadata"] = metadata
		}

		// 6. performs the actions on the memories, creating new ones, updating existing ones, or deleting them.
		functionResultID, err := functionToCall(functionArgs)
		if err != nil {
			utils.DebugPrint(fmt.Sprintf("ERROR calling function %s: %v", functionName, err), m.debug, gc)
			continue
		}

		functionResults = append(functionResults, map[string]interface{}{
			"id":    functionResultID,
			"event": utils.TrimMemorySuffix(functionName),
			"data":  functionArgs["data"],
		})

		// utils.DebugPrint(fmt.Sprintf("Function results: \n%+v\n", functionResults), m.debug)
		// m.telemetry.CaptureEvent("memGo.add.function_call", map[string]interface{}{"memory_id": functionResultID, "function_name": functionName})
	}

	utils.DebugPrint("end", m.debug, gc)
	// 7. returns a list of memories with their IDs, text, and events (ADD, UPDATE, DELETE, or NONE)
	// m.telemetry.CaptureEvent("memGo.add", nil)
	return map[string]interface{}{"message": "ok", "details": functionResults}, nil
}

func (m *Memory) updateMemoryToolWrapper(args map[string]interface{}, gc *gin.Context) (string, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok {
		return "", errors.New("memory_id not found or not a string")
	}
	data, ok := args["data"].(string)
	if !ok {
		return "", errors.New("data not found or not a string")
	}
	return m.updateMemoryTool(memoryID, data, gc)
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
		"id":         memory.Id,
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

	result := utils.MergeMaps(memoryItem, filters)
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
func (m *Memory) Update(memoryID string, data string, gc *gin.Context) (map[string]interface{}, error) {
	// m.telemetry.CaptureEvent("memGo.update", map[string]interface{}{"memory_id": memoryID})
	_, err := m.updateMemoryTool(memoryID, data, gc)
	if err != nil {
		utils.DebugPrint("Error updating memory: "+err.Error(), m.debug, gc)
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

	pacific, err := time.LoadLocation("America/Argentina/Buenos_Aires")
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

func (m *Memory) updateMemoryTool(memoryID string, data string, gc *gin.Context) (string, error) {
	utils.DebugPrint(fmt.Sprintf("Updating memory with memoryID = %s\n", memoryID), m.debug, gc)
	utils.DebugPrint(fmt.Sprintf("with data = %s\n", data), m.debug, gc)

	existingMemory, err := m.vectorStore.Get(memoryID)
	if err != nil {
		return "", fmt.Errorf("error getting existing memory: %w", err)
	}
	if existingMemory == nil {
		return "", fmt.Errorf("memory with ID %s not found", memoryID)
	}

	// prevValue := existingMemory.Payload["data"].(string)
	prevPayload := existingMemory.Payload

	prevValueMap := convertQdrantPayload(prevPayload)

	prevValue := prevValueMap["data"].(string)

	utils.DebugPrint(fmt.Sprintln("old Data: ", prevValue), m.debug, gc)

	newMetadata := make(map[string]interface{})
	newMetadata["data"] = data
	newMetadata["hash"] = prevValueMap["hash"]
	newMetadata["created_at"] = prevValueMap["created_at"]

	hometime, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return "", fmt.Errorf("error loading timezone: %w", err)
	}
	newMetadata["updated_at"] = time.Now().In(hometime).Format(time.RFC3339)

	for _, key := range []string{"user_id", "agent_id", "run_id"} {
		if val, ok := prevValueMap[key]; ok {
			newMetadata[key] = val
		}
	}

	//
	_, embeddings, err := m.embeddingModel.Embed(data)
	if err != nil {
		return "", fmt.Errorf("error embedding data: %w", err)
	}

	// esto inserta el vector
	err = m.vectorStore.Update(memoryID, embeddings, newMetadata)
	if err != nil {
		return "", fmt.Errorf("error updating vector store: %w", err)
	}

	// ESTO HACE UN UPDATE EN LA DB DE SEGUIMIENTO
	// err = m.db.AddHistory(memoryID, &prevValue, data, "UPDATE", newMetadata["created_at"].(*string), newMetadata["updated_at"].(*string), 0)
	// if err != nil {
	// 	utils.DebugPrint(fmt.Sprintf("Error adding history: %v", err), m.debug)
	// }
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

	// prevValue := existingMemory.Payload["data"].(string)
	prevPayload := existingMemory.GetPayload()

	prevValueMap := convertQdrantPayload(prevPayload)

	prevValue := prevValueMap["data"].(string)

	err = m.vectorStore.Delete(memoryID)
	if err != nil {
		return "", fmt.Errorf("error deleting from vector store: %w", err)
	}

	pacific, err := time.LoadLocation("America/Argentina/Buenos_Aires")
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

	StartServer(m)

	// search, err := m.Search("hello", &userId, nil, nil, 5, nil)
	// if err != nil {
	// 	log.Fatalf("Error searching memory: %v", err)
	// }
	// fmt.Printf("Search results: %+v\n", search)

}
