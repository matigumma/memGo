package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
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

	collection, err := client.GetCollectionInfo(context.Background(), config["collection_name"].(string))

	if err != nil {
		panic(err)
	}

	if collection == nil {
		err = client.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: config["collection_name"].(string),
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     uint64(config["embedding_model_dims"].(int)),
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			panic(err)
		}
	}

	return &Qdrant{config: config, client: client}
}

func (q *Qdrant) Insert(vectors [][]float64, ids []string, payloads []map[string]interface{}) error {
	points := make([]*qdrant.PointStruct, len(vectors))
	for i, vector := range vectors {
		float32Vector := make([]float32, len(vector))
		for j, val := range vector {
			float32Vector[j] = float32(val)
		}

		// Convert payloads to qdrant.Value
		convertedPayload := make(map[string]*qdrant.Value)
		for key, value := range payloads[i] {
			switch v := value.(type) {
			case string:
				convertedPayload[key] = qdrant.NewValueString(v)
			case int:
				convertedPayload[key] = qdrant.NewValueInt(int64(v))
			case int64:
				convertedPayload[key] = qdrant.NewValueInt(v)
			case float64:
				convertedPayload[key] = qdrant.NewValueDouble(v)
			case bool:
				convertedPayload[key] = qdrant.NewValueBool(v)
			case []string:
				// Convert []string to qdrant.Value_ListValue
				listValues := make([]*qdrant.Value, len(v))
				for j, str := range v {
					listValues[j] = qdrant.NewValueString(str)
				}
				convertedPayload[key] = qdrant.NewValueList(&qdrant.ListValue{Values: listValues})
			default:
				panic(fmt.Sprintf("Unsupported payload type: %T for key %s", value, key))
			}
		}

		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(ids[i]),
			Vectors: qdrant.NewVectors(float32Vector...),
			Payload: convertedPayload,
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
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case float32:
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case int:
			conditions = append(conditions, qdrant.NewMatchInt(key, int64(v)))
		case int64:
			conditions = append(conditions, qdrant.NewMatchInt(key, v))
		case bool:
			conditions = append(conditions, qdrant.NewMatchBool(key, v))
		case []interface{}:
			// Handle slices by creating a condition for each element
			for _, item := range v {
				switch item := item.(type) {
				case string:
					conditions = append(conditions, qdrant.NewMatchKeyword(key, item))
				case int:
					conditions = append(conditions, qdrant.NewMatchInt(key, int64(item)))
				case int64:
					conditions = append(conditions, qdrant.NewMatchInt(key, item))
				case bool:
					conditions = append(conditions, qdrant.NewMatchBool(key, item))
				default:
					panic(errors.New(fmt.Sprintf("Unsupported slice item type: %T for key %s", item, key)))
				}
			}
		case []string:
			// Handle slices of strings
			for _, item := range v {
				conditions = append(conditions, qdrant.NewMatchKeyword(key, item))
			}
		default:
			panic(errors.New(fmt.Sprintf("Unsupported filter value type: %T for key %s", value, key)))
		}
	}

	// Return the final filter with all conditions
	// OJO TODO: REVISAR PORQUE NO SON TODAS MUST LAS CONDICIONES , EN CASO DE TAGS O ENTITIES PODRIA SER
	return &qdrant.Filter{
		Must: conditions,
		// Should: conditions,
	}
}

func Float32Ptr(f float32) *float32 {
	return &f
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
		ScoreThreshold: Float32Ptr(0.9), // coincidencia con un minimo del ultimo decil
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
		// fmt.Println("Filter:", filter)
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

func (q *Qdrant) SearchWithThreshold(query []float32, limit int, filters map[string]interface{}, scoreThreshold float32) ([]SearchResult, error) {
	limite := uint64(limit)

	qdrantFilters := q._createFilter(filters)

	// Create search points
	searchPoints := &qdrant.QueryPoints{
		CollectionName: q.config["collection_name"].(string),
		Query:          qdrant.NewQuery(query...),
		Limit:          &limite,
		Filter:         qdrantFilters,
		WithPayload:    qdrant.NewWithPayload(true),
		ScoreThreshold: Float32Ptr(scoreThreshold), // coincidencia con un minimo del ultimo decil
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
		// fmt.Println("Filter:", filter)
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
		// fmt.Printf("Processing key: %s, value kind: %T\n", k, v)
		switch kind := v.GetKind().(type) {
		case *qdrant.Value_StringValue:
			result[k] = v.GetStringValue()
		case *qdrant.Value_IntegerValue:
			result[k] = v.GetIntegerValue()
		case *qdrant.Value_DoubleValue:
			result[k] = v.GetDoubleValue()
		case *qdrant.Value_BoolValue:
			result[k] = v.GetBoolValue()
		case *qdrant.Value_ListValue:
			// Handle list values
			listValue := v.GetListValue()
			convertedList := make([]interface{}, len(listValue.Values))
			for i, item := range listValue.Values {
				// Convert each item in the list to its appropriate type
				switch itemKind := item.GetKind().(type) {
				case *qdrant.Value_StringValue:
					convertedList[i] = item.GetStringValue()
				case *qdrant.Value_IntegerValue:
					convertedList[i] = item.GetIntegerValue()
				case *qdrant.Value_DoubleValue:
					convertedList[i] = item.GetDoubleValue()
				case *qdrant.Value_BoolValue:
					convertedList[i] = item.GetBoolValue()
				case *qdrant.Value_NullValue:
					convertedList[i] = nil
				default:
					panic(errors.New(fmt.Sprintf("Unsupported list item kind: %T for key %s", itemKind, k)))
				}
			}
			result[k] = convertedList // Store the converted list
		case *qdrant.Value_StructValue:
			result[k] = v.GetStructValue()
		case *qdrant.Value_NullValue:
			result[k] = nil
		default:
			panic(errors.New(fmt.Sprintf("Unsupported value kind: %T for key %s", kind, k)))
		}
	}
	return result
}

func (q *Qdrant) Get(vectorID string) (*qdrant.RetrievedPoint, error) {
	// Convert the vectorID to a Qdrant PointId
	pointID, err := parsePointID(vectorID)
	if err != nil {
		return nil, fmt.Errorf("invalid vector ID: %v", err)
	}

	// Retrieve the point using the Qdrant client
	/*
		Returns:

		[]*RetrievedPoint: A slice of retrieved points.
		error: An error if the operation fails.
	*/
	points, err := q.client.Get(context.Background(), &qdrant.GetPoints{
		CollectionName: q.config["collection_name"].(string),
		Ids:            []*qdrant.PointId{pointID},
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve point: %v", err)
	}

	// Check if any points were returned
	if len(points) == 0 {
		return nil, fmt.Errorf("no point found with ID: %s", vectorID)
	}

	// Take the first point (since we requested a single point)

	point := points[0]

	return point, nil
}

// Helper function to parse vectorID to Qdrant PointId
func parsePointID(vectorID string) (*qdrant.PointId, error) {
	// Try parsing as a number first
	if numID, err := strconv.ParseUint(vectorID, 10, 64); err == nil {
		// fmt.Printf("Parsed number: %d\n", numID)
		return qdrant.NewIDNum(numID), nil
	}

	if strings.HasPrefix(vectorID, "uuid:") {
		parts := strings.Split(vectorID, ":")
		if len(parts) == 2 {
			real_uuid := strings.Trim(parts[1], "\"")
			// Try parsing as a UUID
			if uuidID, err := uuid.Parse(real_uuid); err == nil {
				// fmt.Printf("Parsed UUID: %s\n", uuidID.String())
				return qdrant.NewID(uuidID.String()), nil
			}
		}
	}

	// If not a number or UUID, treat as a string ID
	if vectorID == "" {
		return nil, errors.New("vectorID cannot be empty")
	}

	// fmt.Printf("Parsed string: %s\n", vectorID)
	return qdrant.NewID(vectorID), nil
}

func (q *Qdrant) List(filters map[string]interface{}, limit int) ([][]SearchResult, error) {
	return nil, errors.New("Qdrant.List not implemented")
}
func (q *Qdrant) Update(vectorID string, vector []float32, payload map[string]interface{}) error {
	pointID, err := parsePointID(vectorID)
	if err != nil {
		return fmt.Errorf("invalid vector ID: %v", err)
	}

	//update vector
	updateVectorResult, err := q.client.UpdateVectors(context.Background(), &qdrant.UpdatePointVectors{
		CollectionName: q.config["collection_name"].(string),
		Points: []*qdrant.PointVectors{
			{
				Id: pointID,
				Vectors: qdrant.NewVectorsMap(map[string]*qdrant.Vector{
					"data": qdrant.NewVector(vector...),
				}),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update vector: %v", err)
	}
	_ = updateVectorResult

	// update payload
	payloadResult, err := q.client.SetPayload(context.Background(), &qdrant.SetPayloadPoints{
		CollectionName: q.config["collection_name"].(string),
		Payload:        qdrant.NewValueMap(payload),
		PointsSelector: qdrant.NewPointsSelector(pointID),
	})
	if err != nil {
		return fmt.Errorf("failed to update payload: %v", err)
	}
	_ = payloadResult

	return nil
}
func (q *Qdrant) Delete(vectorID string) error {
	return errors.New("Qdrant.Delete not implemented")
}
func (q *Qdrant) DeleteCol() error {
	return errors.New("Qdrant.DeleteCol not implemented")
}
