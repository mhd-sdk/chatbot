package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/mhd-sdk/chatbot/model"
	"github.com/ollama/ollama/api"
)

func initHandlers(s *Server) {
	s.fiberServer.Post("/chats", createChat(s))
	s.fiberServer.Post("/chats/:chatID/message", sendMessage(s))
	s.fiberServer.Get("/chats", getChats(s))
}

// create a chat and return the chat id
func createChat(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chat := model.Chat{}
		s.db.Create(&chat)
		return c.JSON(chat)
	}
}

type ChatRequest struct {
	Prompt string `json:"prompt"`
}

func sendMessage(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {

		ID := c.Params("chatID")
		chat := model.Chat{}

		if err := s.db.Preload("Messages").First(&chat, ID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Chat not found"})
		}

		req := new(ChatRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		userMsg := model.Message{
			ChatID:  chat.ID,
			Role:    "user",
			Content: req.Prompt,
		}

		chat.AddMessage(userMsg)
		s.db.Save(&chat)

		ctx := context.Background()
		falseVar := false
		ollamaReq := &api.ChatRequest{
			Model:    "llama3.2",
			Stream:   &falseVar,
			Messages: chat.OllamaMessages(),
		}

		respFunc := func(resp api.ChatResponse) error {
			llmMessage := model.Message{
				ChatID:  chat.ID,
				Content: resp.Message.Content,
				Role:    "assistant",
			}

			chat.AddMessage(llmMessage)
			s.db.Save(&chat)

			c.JSON(chat)

			return nil
		}

		err := s.ollamaClient.Chat(ctx, ollamaReq, respFunc)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get response"})
		}
		return nil

	}
}

func getChats(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var chats []model.Chat
		s.db.Preload("Messages").Find(&chats)
		return c.JSON(chats)
	}
}
