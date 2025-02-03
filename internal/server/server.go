package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	fiberServer *fiber.App
}

func New() *Server {

	fiberServer := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	fiberServer.Use(logger.New(logger.Config{}))

	s := &Server{
		fiberServer: fiberServer,
	}
	initHandlers(s)

	return s
}

func (s *Server) ServeAPI() error {
	slog.Info("Serving api on localhost:3000")
	return s.fiberServer.Listen(":3000")
}
