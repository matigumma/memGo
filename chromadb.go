package main

import "errors"

type ChromaDB struct {
	config map[string]interface{}
}

func NewChromaDB(config map[string]interface{}) *ChromaDB {
	return &ChromaDB{config: config}
}

func (c *ChromaDB) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	return errors.New("ChromaDB.Insert not implemented")
}
func (c *ChromaDB) Search(query []float64, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	return nil, errors.New("ChromaDB.Search not implemented")
}
func (c *ChromaDB) Get(vectorID string) (*SearchResult, error) {
	return nil, errors.New("ChromaDB.Get not implemented")
}
func (c *ChromaDB) List(filters map[string]interface{}, limit int) ([][]SearchResult, error) {
	return nil, errors.New("ChromaDB.List not implemented")
}
func (c *ChromaDB) Update(vectorID string, vector []float64, payload map[string]interface{}) error {
	return errors.New("ChromaDB.Update not implemented")
}
func (c *ChromaDB) Delete(vectorID string) error {
	return errors.New("ChromaDB.Delete not implemented")
}
func (c *ChromaDB) DeleteCol() error {
	return errors.New("ChromaDB.DeleteCol not implemented")
}
