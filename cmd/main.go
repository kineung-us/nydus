package main

import (
	"context"
	"nydus/pkg/env"
	"nydus/pkg/handler"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guiguan/caster"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/helmet/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	app.All("/*", func(c *fiber.Ctx) error {
		url := env.TargetRoot + c.Params("*")
		if err := proxy.Do(c, url); err != nil {
			return err
		}
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	go func() {
		log.Info().Str("Server start", env.Nversion).
			Str("Port", env.ServiceAddress).
			Send()

		log.Debug().
			Str("stage", "env").
			Bool("Debug", env.Debug).
			Str("ServiceAddress", env.ServiceAddress).
			Str("ServiceIP", env.ServiceIP).
			Str("SubscribePubsub", env.SubscribePubsub).
			Str("SubscribeTopic", env.SubscribeTopic).
			Str("PublishPubsub", env.PublishPubsub).
			Str("PublishPubsubTTL", env.PublishPubsubTTL).
			Str("TargetRoot", env.TargetRoot).
			Str("TargetVersion", env.TargetVersion).
			Str("InvokeTimeout", env.InvokeTimeout).
			Str("PublishTimeout", env.PublishTimeout).
			Str("CallbackTimeout", env.CallbackTimeout).
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
