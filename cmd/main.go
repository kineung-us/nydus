package main

import (
	"context"
	"nydus/pkg/env"
	"nydus/pkg/handler"
	"os"
	"os/signal"
	"syscall"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/guiguan/caster"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/helmet/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	app := fiber.New(fiber.Config{
		ServerHeader:             env.Nversion,
		DisableHeaderNormalizing: true,
		DisableStartupMessage:    true,
	})
	app.Use(helmet.New())

	cst := caster.New(context.TODO())

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if env.Debug {
		// /debug/pprof
		app.Use(pprof.New())
		app.Get("/dashboard", monitor.New())

		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// https://v1-rc3.docs.dapr.io/developing-applications/building-blocks/pubsub/howto-publish-subscribe/
	app.Get("/dapr/subscribe", func(c *fiber.Ctx) error {
		sub := []struct {
			Pubsubname string `json:"pubsubname"`
			Topic      string `json:"topic"`
			Route      string `json:"route"`
		}{{
			Pubsubname: env.SubscribePubsub,
			Topic:      env.SubscribeTopic,
			Route:      "/invoke",
		}}
		log.Info().Interface("sub", sub).Send()
		return c.JSON(sub)
	})

	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	app.All("/publish/:target/*", handler.PublishHandler(cst))
	app.Post("/log", handler.LogHandler)
	app.Post("/callback/:id", handler.CallbackHandler(cst))
	app.Post("/invoke", handler.InvokeHandler)

	go func() {
		log.Info().Str("Server start", env.Nversion).
			Str("Port", env.ServiceAddress).
			Send()
		if err := app.Listen(":" + env.ServiceAddress); err != nil {
			log.Panic().Err(err).Send()
		}
	}()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-c
	_ = app.Shutdown()
}
