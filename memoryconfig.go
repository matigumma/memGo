package main

// MemoryConfig - Corresponds to the Python MemoryConfig class
type MemoryConfig struct {
	VectorStore   VectorStoreConfig `json:"vector_store"`
	Llm           LlmConfig         `json:"llm"`
	Embedder      EmbedderConfig    `json:"embedder"`
	HistoryDBPath string            `json:"history_db_path" default:"./history.db"`
}

// NewMemoryConfig creates a new MemoryConfig with default values
func NewMemoryConfig() MemoryConfig {
	return MemoryConfig{
		VectorStore: VectorStoreConfig{
			// Provider options: "qdrant", "chroma", "pgvector"
			Provider: "qdrant",
			Config:   map[string]interface{}{},
		},
		Llm: LlmConfig{
			//Provider of the LLM (e.g., 'ollama', 'openai')
			Provider: "openai",
			Config:   map[string]interface{}{},
		},
		Embedder: EmbedderConfig{
			// Provider options: "openai", "ollama", "huggingface", "azure_openai"
			Provider: "openai",
			Config:   map[string]interface{}{},
		},
		HistoryDBPath: "./history.db", //Path to the history database
	}
}

// Validate validates the MemoryConfig
func (mc *MemoryConfig) Validate() error {
	if err := mc.Embedder.ValidateConfig(); err != nil {
		return err
	}
	if err := mc.Llm.ValidateConfig(); err != nil {
		return err
	}
	if err := mc.VectorStore.ValidateAndCreateConfig(); err != nil {
		return err
	}
	return nil
}
