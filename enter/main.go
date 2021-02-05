package main

import (
	"github.com/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/guiguan/caster"
)

func main() {
	app := fiber.New()
	p := caster.New(nil)

	app.Get("/a/:id", func(c *fiber.Ctx) error {
		ch, ok := p.Sub(nil, 1)
		if !ok {
			return errors.Errorf("invalid event data type")
		}
		defer p.Unsub(ch)

		body := []byte{}

		for m := range ch {
			t := m.(message).ID
			if c.Params("id") == t {
				body = m.(message).Body
				break
			}
		}
		return c.Send(body)
	})

	app.Post("/b/:id", func(c *fiber.Ctx) error {

		p.Pub(message{
			ID:   c.Params("id"),
			Body: c.Body(),
		})
		return c.SendString("send!")
	})

	app.Listen(":3000")
}

type message struct {
	ID   string
	Body []byte
}
