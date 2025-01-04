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
			// Provider options: "openai", "ollama", "groq", "together", "aws_bedrock", "litellm", "azure_openai"
			Provider: "openai",
			Config:   map[string]interface{}{},
		},
		Llm: LlmConfig{
			Provider: "openai",
			Config:   map[string]interface{}{},
		},
		Embedder: EmbedderConfig{
			// Provider options: "openai", "ollama", "huggingface", "azure_openai"
			Provider: "openai",
			Config:   map[string]interface{}{},
		},
		HistoryDBPath: "./history.db",
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
