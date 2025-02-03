package server

import (
	"bufio"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mhd-sdk/chatbot/model"
	"github.com/ollama/ollama/api"
)

func initHandlers(s *Server) {
	s.fiberServer.Post("/chats", createChat(s))
	s.fiberServer.Get("/chats/:chatID", getChat(s))
	s.fiberServer.Post("/chats/:chatID/message", sendMessage(s))
	s.fiberServer.Get("/chats", getChats(s))
}

func getChat(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ID := c.Params("chatID")
		chat := model.Chat{}
		if err := s.db.Preload("Messages").First(&chat, ID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Chat not found"})
		}
		return c.JSON(chat)
	}
}

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

		c.Set("Content-Type", "text/event-stream; charset=utf-8")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// Define the function that will stream the response to the client
		completeResponse := ""
		c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			ctx := context.Background()
			stream := true
			ollamaRequest := &api.ChatRequest{
				Model:    "llama3.2",
				Stream:   &stream,
				Messages: chat.OllamaMessages(),
			}

			respFunc := func(resp api.ChatResponse) error {
				fmt.Fprintf(w, "%s", resp.Message.Content)
				completeResponse += resp.Message.Content
				err := w.Flush()
				if err != nil {
					fmt.Println("Error flushing the writer", err)
				}
				return nil
			}
			err := s.ollamaClient.Chat(ctx, ollamaRequest, respFunc)
			if err != nil {
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get response"})
			}
			llmMessage := model.Message{
				ChatID:  chat.ID,
				Content: completeResponse,
				Role:    "assistant",
			}
			chat.AddMessage(llmMessage)
			s.db.Save(&chat)
		})
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
