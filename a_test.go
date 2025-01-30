package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAddMemoryHandler tests the /v1/memory/add endpoint
func TestAddMemoryHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Create a mock Memory instance
	mockMemory := &Memory{
		// Initialize with necessary fields or mock dependencies
	}

	// Define the route
	router.POST("/v1/memory/add", func(c *gin.Context) {
		addMemoryHandler(c, mockMemory)
	})

	// Create a new HTTP request
	req, _ := http.NewRequest("POST", "/v1/memory/add", nil)
	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Event 0")
}

// TestRetrieveMemoryHandler tests the /v1/memory/retrieve endpoint
func TestRetrieveMemoryHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Create a mock Memory instance
	mockMemory := &Memory{
		// Initialize with necessary fields or mock dependencies
	}

	// Define the route
	router.GET("/v1/memory/retrieve", func(c *gin.Context) {
		retrieveMemoryHandler(c, mockMemory)
	})

	// Create a new HTTP request with query parameters
	req, _ := http.NewRequest("GET", "/v1/memory/retrieve?query=hello&user_id=Matias&agent_id=http", nil)
	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "thoughts")
}

func TestRetrieveMemoryHandlerBadRequest(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Create a mock Memory instance
	mockMemory := &Memory{
		// Initialize with necessary fields or mock dependencies
	}

	// Define the route
	router.GET("/v1/memory/retrieve", func(c *gin.Context) {
		retrieveMemoryHandler(c, mockMemory)
	})

	// Create a new HTTP request without required query parameters
	req, _ := http.NewRequest("GET", "/v1/memory/retrieve", nil)
	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
