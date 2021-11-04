package handler

import (
	"nydus/pkg/body"
	"nydus/pkg/call"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func InvokeHandler(c *fiber.Ctx) error {
	before := time.Now()
	ce := body.CustomEvent{}

	log.Debug().
		Str("stage", "invoke-start").
		Str("body", string(c.Body())).
		Send()

	if err := json.Unmarshal(c.Body(), &ce); err != nil {
		log.Error().Stack().Err(err).Send()
		return fiber.NewError(500, "invokeHandler: CloudEvent Data Unmarchal failed. Err:", err.Error())
	}

	log.Debug().
		Str("TraceID", ce.TraceID).
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
		Str("TraceID", ce.TraceID).
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
		Str("TraceID", ce.TraceID).
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
		Str("TraceID", ce.TraceID).
		Str("service", subTopic).
		Str("version", version).
		Str("route", "/invoke").
		Str("latency", after.Sub(before).String()).
		Interface("request", ce).
		Interface("response", cb).
		Send()

	return c.JSON(fiber.Map{"success": true})
}
