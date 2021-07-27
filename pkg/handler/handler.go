package handler

import (
	"context"
	"strconv"
	"time"

	"nydus/pkg/body"
	"nydus/pkg/call"
	"nydus/pkg/env"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/guiguan/caster"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	root     = env.TargetRoot
	subTopic = env.SubscribeTopic
	version  = env.TargetVersion
	port     = env.ServiceAddress
	IP       = env.ServiceIP
)

func ProxyHendler(c *fiber.Ctx) error {
	url := root + "/" + c.Params("*")
	log.Debug().
		Str("stage", "/*").
		Str("ProxyAddress", url).
		Send()
	if err := proxy.Do(c, url); err != nil {
		log.Panic().Err(err).Send()
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

func InvokeHandler(c *fiber.Ctx) error {
	before := time.Now()
	ce := body.CustomEvent{}

	if err := json.Unmarshal(c.Body(), &ce); err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "invokeHandler: CloudEvent Data Unmarchal failed. Err:", err.Error())
	}

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("service", subTopic).
		Str("route", "/invoke").
		Str("locate", "after-marshal").
		Interface("request", ce).
		Send()

	if err := ce.Data.UpdateHost(root); err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "invokeHandler: UpdateHost Method failed. Err:", err.Error())
	}
	ce.PropTrace()

	out, err := call.RequesttoTarget(ce.Data.Order)
	if err != nil {
		out = &body.ResponseData{
			Body: []byte(err.Error()),
		}
	}

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("service", subTopic).
		Str("route", "/invoke").
		Str("locate", "after-requesttoTarget").
		Interface("request", ce).
		Interface("response", out).
		Send()

	cb := body.Callback{
		Callback: ce.Data.Callback,
		ID:       ce.ID,
		Response: out,
	}

	if err := call.CallbacktoSource(&cb); err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "invokeHandler: CallbackTo Source Err: ", err.Error())
	}

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("service", subTopic).
		Str("route", "/invoke").
		Str("locate", "after-callback").
		Interface("request", ce).
		Interface("response", cb).
		Str("responseBodyString", string(cb.Response.Body.([]byte))).
		Send()

	after := time.Now()

	b, err := body.Unmarshal(cb.Response.Body.([]byte), cb.Response.Headers["Content-Type"])
	if err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "invokeHandler: Response Body Unmarshal Err: ", err.Error())
	}
	cb.Response.Body = b

	log.Info().
		Str("traceid", ce.TraceID).
		Str("service", subTopic).
		Str("version", version).
		Str("route", "/invoke").
		Str("latency", after.Sub(before).String()).
		Interface("request", ce).
		Interface("response", cb).
		Send()

	return c.JSON(fiber.Map{"success": true})
}

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

func PublishHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
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

		ce := body.NewCustomEvent(&pub, getTrace(c), c.Params("target"))

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

		return c.Status(st).Send(rb)
	}
}

func getTrace(c *fiber.Ctx) string {
	tid := ""
	switch {
	case c.Get("traceparent") != "":
		tid = c.Get("traceparent")
	case c.Get("traceid") != "":
		tid = c.Get("traceid")
	default:
	}
	return tid
}
