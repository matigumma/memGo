package main

import (
	"errors"

	"github.com/qdrant/go-client/qdrant"
)

type ChromaDB struct {
	config map[string]interface{}
}

//	func NewChromaDB(config map[string]interface{}) *ChromaDB {
//		return &ChromaDB{config: config}
//	}
func NewChromaDB(config map[string]interface{}) VectorStore {
	return &ChromaDB{config: config}
}

func (c *ChromaDB) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	return errors.New("ChromaDB.Insert not implemented")
}
func (c *ChromaDB) Search(query []float32, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	return nil, errors.New("ChromaDB.Search not implemented")
}
func (c *ChromaDB) SearchWithThreshold(query []float32, limit int, filters map[string]interface{}, scoreThreshold float32) ([]SearchResult, error) {
	return nil, errors.New("ChromaDB.SearchWithThreshold not implemented")
}
func (c *ChromaDB) Get(vectorID string) (*qdrant.RetrievedPoint, error) {
	return nil, errors.New("ChromaDB.Get not implemented")
}
func (c *ChromaDB) List(filters map[string]interface{}, limit int) ([][]SearchResult, error) {
	return nil, errors.New("ChromaDB.List not implemented")
}
func (c *ChromaDB) Update(vectorID string, vector []float32, payload map[string]interface{}) error {
	return errors.New("ChromaDB.Update not implemented")
}
func (c *ChromaDB) Delete(vectorID string) error {
	return errors.New("ChromaDB.Delete not implemented")
}
func (c *ChromaDB) DeleteCol() error {
	return errors.New("ChromaDB.DeleteCol not implemented")
}
