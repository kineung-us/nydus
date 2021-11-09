package main

import (
	"context"
	"nydus/pkg/call"
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
	"github.com/gofiber/helmet/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	daprInit = false

	hthzTimeout = time.Second * time.Duration(env.HealthzTimeout)

	svrReadTO     = env.ServerReadTimeoutSec
	svrWriteTO    = env.ServerWriteTimeoutSec
	svrIdleTO     = env.ServerIdleTimeoutSec
	svrHeaderNorm = env.ServerHeaderNormalizing
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	app := fiber.New(fiber.Config{
		ServerHeader:              env.Nversion,
		Immutable:                 true,
		BodyLimit:                 -1,
		ReadTimeout:               time.Second * time.Duration(svrReadTO),
		WriteTimeout:              time.Second * time.Duration(svrWriteTO),
		IdleTimeout:               time.Second * time.Duration(svrIdleTO),
		DisableDefaultContentType: true,
		DisableHeaderNormalizing:  !svrHeaderNorm,
		DisableStartupMessage:     true,
		ReduceMemoryUsage:         true,
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
	app.Use("/*", handler.DaprInitChk(&daprInit))
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("OK")
	})

	app.All("/publish/:target/*", handler.PublishHandler(cst))
	app.Post("/callback/:id", handler.CallbackHandler(cst))
	app.Post("/invoke", handler.InvokeHandler)
	app.Post("/log", handler.LogHandler)

	app.All("/*", handler.PublishHandler(cst))

	go func() {
		log.Debug().
			Str("stage", "env").
			Bool("Debug", env.Debug).
			Str("ServiceAddress", env.ServiceAddress).
			Str("ServiceIP", env.ServiceIP).
			Str("SubscribePubsub", env.SubscribePubsub).
			Str("SubscribeTopic", env.SubscribeTopic).
			Str("PublishPubsub", env.PublishPubsub).
			Str("PublishPubsubTTL", env.PublishPubsubTTL).
			Str("TargetRoot", env.TargetRoot.String()).
			Str("TargetVersion", env.TargetVersion).
			Int("InvokeTimeout", env.InvokeTimeout).
			Int("PublishTimeout", env.PublishTimeout).
			Int("CallbackTimeout", env.CallbackTimeout).
			Send()

		log.Info().Str("stage", "Target app wait").
			Str("TargetRoot", env.TargetRoot.String()).
			Str("TargetHealthzPath", env.TargetHealthzPath).
			Send()
		for {
			st := call.TargetHealthChk()
			if st < 400 && st >= 200 {
				break
			}
			time.Sleep(hthzTimeout)
		}
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
