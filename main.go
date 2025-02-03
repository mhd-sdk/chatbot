package main

import (
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Post("/upload", func(c *fiber.Ctx) error {
		// Open a file to write the streamed data
		file, err := os.Create("uploaded_file")
		if err != nil {
			return err
		}
		defer file.Close()

		// Stream the request body to the file
		_, err = io.Copy(file, c.Request().BodyStream())
		if err != nil {
			return err
		}

		return c.SendString("File uploaded successfully")
	})

	log.Fatal(app.Listen(":3000"))
}
