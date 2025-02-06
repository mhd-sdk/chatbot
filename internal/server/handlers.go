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
	s.fiberServer.Post("/chats/:userID", createChat(s))
	s.fiberServer.Post("/chats/:chatID/message", sendMessage(s))
	s.fiberServer.Get("/chats/:chatID", getChat(s))
	s.fiberServer.Get("/chats/user/:userID", getChats(s))
	s.fiberServer.Put("/chats/:chatID", renameChat(s))    // Handler pour renommer un chat
	s.fiberServer.Delete("/chats/:chatID", deleteChat(s)) // Handler pour supprimer un chat
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
		var request struct {
			Name string `json:"name"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		chat := model.Chat{
			UserID: c.Params("userID"),
			Name:   request.Name,
		}
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
				fmt.Println(err)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get response", "ollama error": err.Error()})
			}
			llmMessage := model.Message{
				ChatID:  chat.ID,
				Content: completeResponse,
				Role:    "assistant",
			}
			chat.AddMessage(llmMessage)
			s.db.Save(&chat)
		})
		return c.SendStatus(fiber.StatusOK)
	}
}

func getChats(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var chats []model.Chat
		s.db.Preload("Messages").Where("user_id = ?", c.Params("userID")).Find(&chats)
		return c.JSON(chats)
	}
}

func renameChat(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ID := c.Params("chatID")
		var request struct {
			Name string `json:"name"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		chat := model.Chat{}
		if err := s.db.First(&chat, ID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Chat not found"})
		}

		chat.Name = request.Name
		s.db.Save(&chat)
		return c.JSON(chat)
	}
}

func deleteChat(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ID := c.Params("chatID")
		if err := s.db.Delete(&model.Chat{}, ID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Chat not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
