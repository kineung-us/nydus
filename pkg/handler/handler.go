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

	root        = env.TargetRoot.String()
	subTopic    = env.SubscribeTopic
	serviceName = env.SubscribeTopic
	version     = env.TargetVersion
	port        = env.ServiceAddress
	IP          = env.ServiceIP
)

func CallbackHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		log.Debug().
			Str("stage", "callback-start").
			Str("body", string(c.Body())).
			Send()

		hd := map[string]string{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			hd[string(key)] = string(value)
		})

		b, err := body.Unmarshal(c.Body(), getContentType(c))
		if err != nil {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(500, "callbackHandler: Body Marchal failed. Err: ", err.Error())
		}

		m := body.Message{
			ID:      c.Params("id"),
			Status:  getStatus(c),
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
	log.Debug().
		Str("stage", "log-start").
		Str("body", string(c.Body())).
		Send()

	b, err := body.Unmarshal(c.Body(), getContentType(c))
	if err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "logHandler: Body Json Marchal failed. Err: ", err.Error())
	}

	log.Info().
		Str("service", subTopic).
		Str("version", version).
		Str("route", c.OriginalURL()).
		Str("contentType", getContentType(c)).
		Interface("request", b).
		Send()
	return c.SendStatus(204)
}

func DaprInitChk(d *bool) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		log.Debug().
			Str("stage", "daprinit-start").
			Str("body", string(c.Body())).
			Send()

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

func getStatus(c *fiber.Ctx) string {
	return c.Get("status") + c.Get("Status")
}

func getContentType(c *fiber.Ctx) string {
	return c.Get("content-type") + c.Get("Content-Type")
}
