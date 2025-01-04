package main

import "errors"

type Qdrant struct {
	config map[string]interface{}
}

func NewQdrant(config map[string]interface{}) *Qdrant {
	return &Qdrant{config: config}
}

func (q *Qdrant) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	return errors.New("Qdrant.Insert not implemented")
}
func (q *Qdrant) Search(query []float64, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	return nil, errors.New("Qdrant.Search not implemented")
}
func (q *Qdrant) Get(vectorID string) (*SearchResult, error) {
	return nil, errors.New("Qdrant.Get not implemented")
}
func (q *Qdrant) List(filters map[string]interface{}, limit int) ([][]SearchResult, error) {
	return nil, errors.New("Qdrant.List not implemented")
}
func (q *Qdrant) Update(vectorID string, vector []float64, payload map[string]interface{}) error {
	return errors.New("Qdrant.Update not implemented")
}
func (q *Qdrant) Delete(vectorID string) error {
	return errors.New("Qdrant.Delete not implemented")
}
func (q *Qdrant) DeleteCol() error {
	return errors.New("Qdrant.DeleteCol not implemented")
}
