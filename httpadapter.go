package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock function to simulate streaming events
func mockStream(c *gin.Context) {
	c.Stream(func(w io.Writer) bool {
		// Simulate sending events every second
		for i := 0; i < 5; i++ {
			time.Sleep(1 * time.Second)
		}
		return false
	})
}

// Handler for /v1/memory/add
func addMemoryHandler(c *gin.Context, m *Memory) {
	var requestBody struct {
		Text    string `json:"text"`
		UserID  string `json:"user_id"`
		AgentID string `json:"agent_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Start streaming
	c.Writer.Header().Set("Content-Type", "text/event-stream")

	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	/*
		// text := "Hola, me contactó un posible cliente que necesita implementar un chatboot que participando de un grupo de whatsapp analice las conversaciones para encontrar cierta información y después al encontrarse con ciertos parámetros contacte por whatsapp a números que se encuentran en la conversación misma y le mande un mensaje y tal vez le permita ingresar información que debe ser persistida en una base de datos. En ITR podemos hacer este desarrollo, pero no me cierra el tamaño del cliente / posibilidades económicas. Si a alguien le interesa contácteme por privado para ponerlo en contacto con el cliente"
		// text := "vengo acá a recordarles que mañana a las 17 hacemos el brainstorming y reunión de encuentro, con los que puedan sumarse."
		// text := "Buen día, consultita en el grupo ¿han socializado algún material sobre ingeniería de prompts?"
		// text += "Para darles contexto estoy preparando un documento de prompts para que le sirva a 3 equipos (copy,diseño y comtent) para la empresa en la que trabajo. De modo que quería tener otros recursos bibliográficas para ampliar el material"
		// text := "Gus, creo que podemos arrancar con una esa semana, y después a fin de enero la continuamos con una más.. no creo que con una sola reunión semejante profusión de ideas se pueda hacer converger de una"
		// text := "Hola, buen dia"
		// text := "hay que quedar un monto para 10 siguientes y te envió por crypto. El anterior fueron $100 equivalentes en crypto por 10 adicionales, lo repetimos?"
	*/

	res, err := m.Add(
		requestBody.Text,     // data
		&requestBody.UserID,  // user_id
		&requestBody.AgentID, // agent_id
		nil,                  // run_id
		nil,                  // metadata
		nil,                  // filters
		nil,                  // custom prompt
		c,                    // gin context
	)
	if err != nil {
		log.Fatalf("Error in Memory.Add: %v", err)
	}

	_ = res

}

// Handler for /v1/memory/retrieve
func retrieveMemoryHandler(c *gin.Context, m *Memory) {
	query := c.Query("query")
	userID := c.Query("user_id")
	agentID := c.Query("agent_id")
	// runID := c.Query("run_id")
	if query == "" || userID == "" || agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query, userID, agentID parameter is required"})
		return
	}

	// Perform a search to test the Search function
	// searchQuery := "ingeniería de prompts"
	searchResults, err := m.Search(query, &userID, &agentID, nil, 5, nil)
	if err != nil {
		log.Fatalf("Error searching memory: %v", err)
	}
	// fmt.Printf("Search results for query '%s': %+v\n", searchQuery, searchResults)

	Thoughts := []string{}
	for i, result := range searchResults {
		relatedThoughts := []string{}
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("Result %d:\n", i+1))
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("Memory: %s\n", result["memory"]))
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("  ID: %s\n", result["id"]))
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("  Score: %f\n", result["score"]))
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("  Created At: %s\n", result["created_at"]))
		relatedThoughts = append(relatedThoughts, fmt.Sprintf("  Updated At: %s\n", result["updated_at"]))
		relatedThoughts = append(relatedThoughts, "  Payload:\n")
		for key, value := range result["metadata"].(map[string]interface{}) {
			relatedThoughts = append(relatedThoughts, fmt.Sprintf("    %s: %v\n", key, value))
		}

		Thoughts = append(Thoughts, relatedThoughts...)
	}

	c.JSON(http.StatusOK, gin.H{"thoughts": Thoughts})
}

// StartServer initializes and starts the Gin server
func StartServer(m *Memory) {
	r := gin.Default()
	// Define routes
	r.POST("/v1/memory/add", func(c *gin.Context) {
		addMemoryHandler(c, m)
	})
	r.GET("/v1/memory/retrieve", func(c *gin.Context) {
		retrieveMemoryHandler(c, m)
	})

	// Start the server
	r.Run(":8080")
}
