package main

import (
	"fmt"
	"path/filepath"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore - Interface for Vector Stores (already defined, ensuring it's here for context)
type VectorStore interface {
	Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error
	Search(query []float32, limit int, filters map[string]interface{}) ([]SearchResult, error)
	Get(vectorID string) (*qdrant.RetrievedPoint, error)
	List(filters map[string]interface{}, limit int) ([][]SearchResult, error)
	Update(vectorID string, vector []float32, payload map[string]interface{}) error
	Delete(vectorID string) error
	DeleteCol() error
}

// VectorStoreConfig -
type VectorStoreConfig struct {
	Provider string                 `json:"provider" default:"qdrant"`
	Config   map[string]interface{} `json:"config"`
}

// VectorStoreProviderConfig - Configuration specific to the vector store provider
type VectorStoreProviderConfig struct {
	CollectionName string `json:"collection_name"`
}

// ValidateAndCreateConfig validates the VectorStoreConfig and potentially creates a specific config struct
func (vsc *VectorStoreConfig) ValidateAndCreateConfig() error {
	providerConfigs := map[string]string{
		"qdrant":   "QdrantConfig",   // Placeholder - Define these if needed
		"chroma":   "ChromaDbConfig", // Placeholder
		"pgvector": "PGVectorConfig", // Placeholder
	}

	if _, ok := providerConfigs[vsc.Provider]; !ok {
		return fmt.Errorf("unsupported vector store provider: %s", vsc.Provider)
	}

	// Handle default path if config is nil or a map without "path"
	if vsc.Config == nil {
		vsc.Config = map[string]interface{}{"path": filepath.Join("/tmp", vsc.Provider)}
	} else if _, ok := vsc.Config["path"]; !ok {
		if vsc.Config == nil {
			vsc.Config = make(map[string]interface{})
		}
		vsc.Config["path"] = filepath.Join("/tmp", vsc.Provider)
	}

	return nil
}
