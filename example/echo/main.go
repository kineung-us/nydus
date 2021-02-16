package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Fiber instance
	app := fiber.New()

	// Routes
	app.Post("/:time", hello)

	// Start server
	log.Fatal(app.Listen(":3000"))
}

// Handler
func hello(c *fiber.Ctx) error {
	log.Printf(string(c.Body()))
	to, _ := time.ParseDuration(c.Params("time") + "s")
	fmt.Println(to)
	time.Sleep(to)
	return c.JSON(fiber.Map{"hello": "test"})
}
