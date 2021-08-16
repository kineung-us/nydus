package handler

import (
	"context"
	"nydus/pkg/body"
	"nydus/pkg/call"
	"nydus/pkg/env"

	"github.com/gofiber/fiber/v2"
	"github.com/guiguan/caster"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	root        = env.TargetRoot
	subTopic    = env.SubscribeTopic
	serviceName = env.SubscribeTopic
	version     = env.TargetVersion
	port        = env.ServiceAddress
	IP          = env.ServiceIP
)

func CallbackHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		hd := map[string]string{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			hd[string(key)] = string(value)
		})

		b, err := body.Unmarshal(c.Body(), c.Get("Content-Type"))
		if err != nil {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(500, "callbackHandler: Body Json Marchal failed. Err: ", err.Error())
		}

		m := body.Message{
			ID:      c.Params("id"),
			Status:  c.Get("status"),
			Headers: hd,
			Body:    b,
		}

		if ok := cst.TryPub(m); !ok {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(781, "callbackHandler: Caster delivery failed")
		}
		return c.SendStatus(204)
	}
}

func LogHandler(c *fiber.Ctx) error {
	b, err := body.Unmarshal(c.Body(), c.Get("Content-Type"))
	if err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "logHandler: Body Json Marchal failed. Err: ", err.Error())
	}

	log.Info().
		Str("service", subTopic).
		Str("version", version).
		Str("route", c.OriginalURL()).
		Str("contentType", c.Get("Content-Type")).
		Interface("request", b).
		Send()
	return c.SendStatus(204)
}

func DaprInitChk(d *bool) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log.Debug().Str("func", "DaprInitChk").
			Bool("daprinit", *d).Send()

		ctx := context.Background()
		ctx = context.WithValue(ctx, "daprChk", *d)
		c.SetUserContext(ctx)

		if *d {
			return c.Next()
		}
		chk := call.DaprHealthChk()
		d = &chk
		return c.Next()
	}
}
