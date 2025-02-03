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

func float32Ptr(f float32) *float32 {
	return &f
}

// Handler for /v1/memory/retrieve
func retrieveMemoryHandler(c *gin.Context, m *Memory) {
	// Parse JSON
	var json struct {
		Query   string `json:"query" binding:"required"`
		UserID  string `json:"user_id" binding:"required"`
		AgentID string `json:"agent_id" binding:"required"`
	}

	if err := c.Bind(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	query := json.Query
	userID := json.UserID
	agentID := json.AgentID

	fmt.Printf("Query %s, user_id %s, agent_id %s\n", query, userID, agentID)

	// declaro busqueda con un threshold  muy permisivo
	searchResults, err := m.Search(query, &userID, &agentID, nil, 5, nil, float32Ptr(0.8))
	if err != nil {
		log.Fatalf("Error searching memory: %v", err)
	}
	fmt.Printf("Search results for query : %+v\n", searchResults)

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

	// Disable CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Define routes
	r.POST("/v1/memory/add", func(c *gin.Context) {
		addMemoryHandler(c, m)
	})
	r.POST("/v1/memory/retrieve", func(c *gin.Context) {
		retrieveMemoryHandler(c, m)
	})
	r.POST("/v1/memory/history", func(c *gin.Context) {
		historyMemoryHandler(c, m)
	})

	// Start the server
	r.Run(":8080")
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Reactions struct {
		Likes    int `json:"likes"`
		Dislikes int `json:"dislikes"`
	} `json:"reactions"`
	Saved bool `json:"saved"`
}

// Handler for /v1/memory/history
func historyMemoryHandler(c *gin.Context, m *Memory) {
	messages := []Message{
		{
			ID:        "1",
			Role:      "user",
			Content:   "Can you help me optimize my React application?",
			Timestamp: time.Date(2024, 3, 10, 10, 30, 0, 0, time.UTC),
			Reactions: struct {
				Likes    int `json:"likes"`
				Dislikes int `json:"dislikes"`
			}{Likes: 2, Dislikes: 0},
			Saved: true,
		},
		{
			ID:        "2",
			Role:      "assistant",
			Content:   "Of course! There are several ways to optimize a React application. Some key areas to focus on include:\n\n1. Implementing proper memo and useMemo usage\n2. Optimizing bundle size\n3. Implementing code splitting\n4. Using proper key props in lists\n\nWhich area would you like to explore first?",
			Timestamp: time.Date(2024, 3, 10, 10, 30, 30, 0, time.UTC),
			Reactions: struct {
				Likes    int `json:"likes"`
				Dislikes int `json:"dislikes"`
			}{Likes: 3, Dislikes: 0},
			Saved: false,
		},
		{
			ID:        "3",
			Role:      "user",
			Content:   "Let's start with code splitting. How can I implement it effectively?",
			Timestamp: time.Date(2024, 3, 11, 15, 45, 0, 0, time.UTC),
			Reactions: struct {
				Likes    int `json:"likes"`
				Dislikes int `json:"dislikes"`
			}{Likes: 1, Dislikes: 0},
			Saved: false,
		},
		{
			ID:        "4",
			Role:      "assistant",
			Content:   "Code splitting in React can be implemented using dynamic imports and React.lazy(). Here's a basic example:\n\n```jsx\nconst MyComponent = React.lazy(() => import('./MyComponent'));\n\nfunction App() {\n  return (\n    <Suspense fallback={<Loading />}> \n      <MyComponent />\n    </Suspense>\n  );\n}\n```\n\nThis will load MyComponent only when it's needed. Would you like to see more advanced patterns?",
			Timestamp: time.Date(2024, 3, 11, 15, 46, 0, 0, time.UTC),
			Reactions: struct {
				Likes    int `json:"likes"`
				Dislikes int `json:"dislikes"`
			}{Likes: 4, Dislikes: 0},
			Saved: true,
		},
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
