package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Fiber instance
	app := fiber.New()

	// Routes
	app.Post("/", hello)

	// Start server
	log.Fatal(app.Listen(":3000"))
}

// Handler
func hello(c *fiber.Ctx) error {
	log.Printf(string(c.Body()))
	time.Sleep(10 * time.Second)
	return c.SendString("hi")
}
