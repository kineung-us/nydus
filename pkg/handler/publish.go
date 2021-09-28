package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"nydus/pkg/body"
	"nydus/pkg/call"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/guiguan/caster"
	"github.com/rs/zerolog/log"
)

func PublishHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		log.Debug().Str("func", "PublishHandler").
			Bool("daprinit", ctx.Value("daprChk").(bool)).Send()
		if !ctx.Value("daprChk").(bool) {
			log.Debug().Str("func", "ProxyHandler").Send()
			url := root + "/" + c.Params("*")
			if err := proxy.Do(c, url); err != nil {
				return err
			}
			// Remove Server header from response
			c.Response().Header.Del(fiber.HeaderServer)
			return nil
		}

		log.Debug().Str("func", "PublishHandler").Send()
		before := time.Now()

		// set headers
		hd := map[string]string{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			hd[string(key)] = string(value)
		})

		log.Debug().
			Str("traceid", getTrace(c)).
			Str("service", subTopic).
			Str("route", c.OriginalURL()).
			Str("locate", "publish-start").
			Str("requestMethod", c.Method()).
			Str("contentType", c.Get("Content-Type")).
			Str("requestBody", string(c.Body())).
			Send()

		b, err := body.Unmarshal(c.Body(), c.Get("Content-Type"))
		if err != nil {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(500, "publishHandler: Request Body Unmarchal failed. Err: ", err.Error())
		}

		r := body.RequestedData{
			Method:  c.Method(),
			URL:     c.BaseURL() + c.OriginalURL(),
			Headers: hd,
			Body:    b,
		}

		pub := body.PublishData{
			Order:    &r,
			Callback: "http://" + IP + ":" + port,
		}

		ce := body.NewCustomEvent(&pub, getTrace(c), getTarget(c))

		log.Debug().
			Str("traceid", ce.TraceID).
			Str("service", subTopic).
			Str("route", c.OriginalURL()).
			Interface("request", ce).
			Send()

		// https://v1-rc3.docs.dapr.io/reference/api/pubsub_api/#http-response-codes
		if err := call.Publishrequestevent(ce); err != nil {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(500, "publishHandler: Fail to publish event. Err: ", err.Error())
		}

		ch, ok := cst.Sub(context.TODO(), 1)
		if !ok {
			// https://docs.gofiber.io/api/fiber#newerror
			return fiber.NewError(782, "publishHandler: Caster subscription failed")
		}
		defer cst.Unsub(ch)

		rm := body.Message{}

		for m := range ch {
			t := m.(body.Message).ID
			if ce.ID.String() == t {
				rm = m.(body.Message)
				break
			}
		}
		for k, v := range rm.Headers {
			c.Set(k, v)
		}
		st, _ := strconv.Atoi(rm.Status)

		rb, err := body.Marshal(rm.Body, rm.Headers["Content-Type"])
		if err != nil {
			log.Error().Stack().Err(err).Send()
			return fiber.NewError(500, "publishHandler: Response Body Json Marchal failed. Err: ", err.Error())
		}

		after := time.Now()
		log.Info().
			Str("traceid", ce.TraceID).
			Str("service", subTopic).
			Str("version", version).
			Str("route", c.OriginalURL()).
			Str("latency", after.Sub(before).String()).
			Interface("request", ce).
			Interface("response", rm).
			Send()

		log.Debug().Int("Status", st).Send()
		return c.Status(st).Send(rb)
	}
}

func getTrace(c *fiber.Ctx) string {
	return c.Get("traceparent") + c.Get("traceid")
}

func getTarget(c *fiber.Ctx) string {
	tar := c.Params("target")
	if !strings.HasPrefix(c.Path(), "/publish") {
		tar = serviceName
	}
	return tar
}
