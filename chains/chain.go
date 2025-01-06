package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type Chain struct{}

func NewChain() *Chain {
	return &Chain{}
}

func (c *Chain) MEMORY_DEDUCTION() (map[string][]string, error) {

	model := "gpt-3.5-turbo"
	// model := "gpt-4o-mini"
	// model := "gpt-4o"

	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		log.Panic(err)
	}
	prompt := prompts.NewPromptTemplate(
		`Deduce los hechos, preferencias y recuerdos del texto proporcionado.
			Solo devuelve los hechos, preferencias y recuerdos en viñetas:
			Texto en lenguaje natural: {{.text}}
			Detalles del usuario/agente: {{.details}}
	
			Restricciones para deducir hechos, preferencias y recuerdos:
			- Los hechos, preferencias y recuerdos deben ser concisos e informativos.
			- No comiences con "A la persona le gusta la Pizza". En su lugar, comienza con "Le gusta la Pizza".
			- No recuerdes los detalles del usuario/agente proporcionados. Solo recuerda los hechos, preferencias y recuerdos.
			- Responde en el mismo idioma del texto.
			- Respuesta en formato JSON con una clave como "facts" y el valor correspondiente será una lista de cadenas.
			- No uses comillas invertidas para el formato JSON.
	
			ejemplo:
			{
				"facts": [
					"Le gusta la Pizza",
					"Prefiere comer en un restaurante"
				]
			}
	
			¿Hechos, preferencias y recuerdos deducidos:?`,
		[]string{"text", "details"},
	)
	llmChain := chains.NewLLMChain(llm, prompt)

	// If a chain only needs one input we can use Run to execute it.
	// We can pass callbacks to Run as an option, e.g:
	//   chains.WithCallback(callbacks.StreamLogHandler{})
	ctx := context.Background()
	out, err := chains.Call(ctx, llmChain, map[string]any{
		"text":    "Jugando a la pelota con los chicos ayer me lesione la pierna izquieda y no voy a poder jugar por 6 meses",
		"details": `{"user": "Ricky", "agent": "ChatGPT", "today": "2023-10-26"}`,
	})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Using model: ", model)

	fmt.Println(out["text"])

	parsedOutput, ok := out["text"].(string)
	if !ok {
		log.Panic("Failed to parse output text")
	}

	var result map[string][]string
	err = json.Unmarshal([]byte(parsedOutput), &result)
	if err != nil {
		log.Panic("Error parsing JSON: ", err)
	}

	fmt.Println("Parsed facts:")
	for _, fact := range result["facts"] {
		fmt.Println("-", fact)
	}

	return result
}
