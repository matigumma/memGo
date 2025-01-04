package main

// this is an example of how to validate a struct using

// import "github.com/go-playground/validator/v10"

// type MemoryItem struct {
// 	ID        string                 `json:"id" validate:"required"`
// 	Memory    string                 `json:"memory" validate:"required"`
// 	Hash      *string                `json:"hash,omitempty"`
// 	Metadata  map[string]interface{} `json:"metadata,omitempty"`
// 	Score     *float64               `json:"score,omitempty"`
// 	CreatedAt *string                `json:"created_at,omitempty"`
// 	UpdatedAt *string                `json:"updated_at,omitempty"`
// }

// // ValidateMemoryItem uses the validator library to validate the MemoryItem struct
// func ValidateMemoryItem(item MemoryItem) error {
// 	validate := validator.New()
// 	return validate.Struct(item)
// }

// func main() {
// 	item := MemoryItem{
// 		// ... populate with data ...
// 		ID: "", // Intentionally missing to trigger validation error
// 	}

// 	err := ValidateMemoryItem(item)
// 	if err != nil {
// 		// Handle validation errors
// 		println("Validation error:", err.Error())
// 	}
// }
