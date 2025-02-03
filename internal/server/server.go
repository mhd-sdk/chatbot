package server

import (
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/mhd-sdk/chatbot/model"
	"github.com/ollama/ollama/api"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Server struct {
	fiberServer  *fiber.App
	ollamaClient *api.Client
	db           *gorm.DB
}

func New() *Server {

	fiberServer := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	fiberServer.Use(logger.New(logger.Config{}))
	fiberServer.Use(cors.New())
	dbURL := os.Getenv("DB_URL")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")

	dsn := dbUser + ":" + dbPwd + "@tcp(" + dbURL + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(model.Chat{}, model.Message{})

	ollamaClient, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		ollamaClient: ollamaClient,
		fiberServer:  fiberServer,
		db:           db,
	}
	initHandlers(s)

	return s
}

func (s *Server) ServeAPI() error {
	slog.Info("Serving api on localhost:3000")
	return s.fiberServer.Listen(":3000")

}
