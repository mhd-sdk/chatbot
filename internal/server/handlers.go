package server

import (
	"github.com/gofiber/fiber/v2"
)

func initHandlers(s *Server) {
	s.fiberServer.Get("/helloworld", getScans(s))
}

func getScans(s *Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	}
}
