package chains

import (
	"context"
	"fmt"

	"github.com/matigumma/memGo/utils"
	"github.com/tmc/langchaingo/llms"
)

// Function to extract text content from MessageContent
func GetTextContent(msg llms.MessageContent) []string {
	var texts []string
	for _, part := range msg.Parts {
		if textPart, ok := part.(llms.TextContent); ok {
			texts = append(texts, textPart.Text)
		}
	}
	return texts
}

func (c *Chain) parseLlmsMessagesContent(messages []llms.MessageContent) string {
	var result string
	for _, msg := range messages {
		textContent := GetTextContent(msg)
		result += fmt.Sprintf("Role: %s, Content: %v\n", msg.Role, textContent)
	}
	return result
}

func (c *Chain) debugPrint(message string) {
	if c.gc != nil {
		utils.DebugPrint(message, c.debug, c.gc)
	}
	// if c.debug {
	// 	fmt.Println("DEBUG:::", message)
	// 	fmt.Println("") // print a separated line
	// }
}

// updateMessageHistory updates the message history with the assistant's
// response and requested tool calls.
func updateMessageHistory(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	respchoice := resp.Choices[0]

	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	return append(messageHistory, assistantResponse)
}

func (c *Chain) executeToolCalls(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	fmt.Println("Executing", len(resp.Choices[0].ToolCalls), "tool calls")
	for _, toolCall := range resp.Choices[0].ToolCalls {
		switch toolCall.FunctionCall.Name {
		default:
			c.debugPrint("Unsupported tool: " + toolCall.FunctionCall.Name)
			c.debugPrint("Call: " + fmt.Sprintf("%+v", toolCall.FunctionCall.Arguments))
		}
	}

	return messageHistory
}
