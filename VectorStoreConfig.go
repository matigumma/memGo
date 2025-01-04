package main

import (
	"fmt"
	"path/filepath"
)

// VectorStoreConfig - Corresponds to mem0.configs.base.VectorStoreConfig
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
	} else if configMap, ok := vsc.Config.(map[string]interface{}); ok {
		if _, pathExists := configMap["path"]; !pathExists {
			configMap["path"] = filepath.Join("/tmp", vsc.Provider)
			vsc.Config = configMap
		}
	} else {
		// If config is not a map, assume it's already the correct specific config struct
		// You might want to add more specific type checking here if necessary
	}

	return nil
}
