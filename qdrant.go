package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type Qdrant struct {
	config map[string]interface{}
	client *qdrant.Client
}

func NewQdrant(config map[string]interface{}) VectorStore {
	/*
		original function arguments:
		Args:
				collection_name (str): Name of the collection.
				embedding_model_dims (int): Dimensions of the embedding model.
				client (QdrantClient, optional): Existing Qdrant client instance. Defaults to None.
				host (str, optional): Host address for Qdrant server. Defaults to None.
				port (int, optional): Port for Qdrant server. Defaults to None.
				path (str, optional): Path for local Qdrant database. Defaults to None.
				url (str, optional): Full URL for Qdrant server. Defaults to None.
				api_key (str, optional): API key for Qdrant server. Defaults to None.
				on_disk (bool, optional): Enables persistent storage. Defaults to False.
	*/
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	if err != nil {
		panic(err)
	}

	client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: config["collection_name"].(string),
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(config["embedding_model_dims"].(int)),
			Distance: qdrant.Distance_Cosine,
		}),
	})

	return &Qdrant{config: config, client: client}
}

func (q *Qdrant) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	points := make([]*qdrant.PointStruct, len(vectors))
	for i, vector := range vectors {
		float32Vector := make([]float32, len(vector))
		for j, val := range vector {
			float32Vector[j] = float32(val)
		}
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(ids[i]),
			Vectors: qdrant.NewVectors(float32Vector...),
			Payload: qdrant.NewValueMap(payloads[i]),
		}
	}

	q.client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: q.config["collection_name"].(string),
		Points:         points,
	})
	return nil
}

func (q *Qdrant) _createFilter(filters map[string]interface{}) *qdrant.Filter {
	if len(filters) == 0 {
		return nil
	}

	// Create conditions slice to hold all filter conditions
	conditions := make([]*qdrant.Condition, 0, len(filters))

	// Convert each filter key-value pair to a Qdrant condition
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			conditions = append(conditions, qdrant.NewMatchKeyword(key, v))
		case float64:
			// Convert float64 to int64 for NewMatchInt
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case float32:
			// Add support for float32
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case int:
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case int64:
			// Add direct support for int64
			conditions = append(conditions, qdrant.NewMatchInt(key, v))
		case bool:
			conditions = append(conditions, qdrant.NewMatchBool(key, v))
		}
	}

	// Return the final filter with all conditions
	return &qdrant.Filter{
		Must: conditions,
	}
}

func (q *Qdrant) Search(query []float32, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	limite := uint64(limit)

	qdrantFilters := q._createFilter(filters)

	// Create search points
	searchPoints := &qdrant.QueryPoints{
		CollectionName: q.config["collection_name"].(string),
		Query:          qdrant.NewQuery(query...),
		Limit:          &limite,
		Filter:         qdrantFilters,
		WithPayload:    qdrant.NewWithPayload(true),
		// Payload and vector in the result:
		// WithVectors:    qdrant.NewWithVectors(true),
	}

	// Add filter if provided
	if filters != nil {
		// Convert filters to Qdrant filter format
		// Note: You'll need to implement _createFilter helper method
		// to convert the filters map to Qdrant filter structure
		filter := q._createFilter(filters)
		searchPoints.Filter = filter
	}

	// Perform the search
	results, err := q.client.Query(context.Background(), searchPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// Convert results to SearchResult format
	searchResults := make([]SearchResult, len(results))
	for i, hit := range results {
		searchResults[i] = SearchResult{
			ID:      hit.Id.String(),
			Score:   float64(hit.Score),
			Payload: convertQdrantPayload(hit.Payload),
		}
	}

	return searchResults, nil
}

// Add this helper function
func convertQdrantPayload(payload map[string]*qdrant.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range payload {
		switch {
		case v.GetKind().(*qdrant.Value_StringValue) != nil:
			result[k] = v.GetStringValue()
		case v.GetKind().(*qdrant.Value_IntegerValue) != nil:
			result[k] = v.GetIntegerValue()
		case v.GetKind().(*qdrant.Value_DoubleValue) != nil:
			result[k] = v.GetDoubleValue()
		case v.GetKind().(*qdrant.Value_BoolValue) != nil:
			result[k] = v.GetBoolValue()
		case v.GetKind().(*qdrant.Value_ListValue) != nil:
			result[k] = v.GetListValue()
		case v.GetKind().(*qdrant.Value_StructValue) != nil:
			result[k] = v.GetStructValue()
		case v.GetKind().(*qdrant.Value_NullValue) != nil:
			result[k] = nil
		}
	}
	return result
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
