package main

import "errors"

type PGVector struct {
	config map[string]interface{}
}

//	func NewPGVector(config map[string]interface{}) *PGVector {
//		return &PGVector{config: config}
//	}
func NewPGVector(config map[string]interface{}) VectorStore {
	return &PGVector{config: config}
}

func (p *PGVector) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	return errors.New("PGVector.Insert not implemented")
}
func (p *PGVector) Search(query []float32, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	return nil, errors.New("PGVector.Search not implemented")
}
func (p *PGVector) Get(vectorID string) (*SearchResult, error) {
	return nil, errors.New("PGVector.Get not implemented")
}
func (p *PGVector) List(filters map[string]interface{}, limit int) ([][]SearchResult, error) {
	return nil, errors.New("PGVector.List not implemented")
}
func (p *PGVector) Update(vectorID string, vector []float64, payload map[string]interface{}) error {
	return errors.New("PGVector.Update not implemented")
}
func (p *PGVector) Delete(vectorID string) error {
	return errors.New("PGVector.Delete not implemented")
}
func (p *PGVector) DeleteCol() error {
	return errors.New("PGVector.DeleteCol not implemented")
}
