package main

import (
	jsonstd "encoding/json"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/guiguan/caster"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/helmet/v2"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	version  = "dev"
	debug, _ = strconv.ParseBool(getEnvVar("DEBUG", "false"))

	serviceAddress = getEnvVar("APP_PORT", "5000")
	myIP           = getEnvVar("MY_POD_IP", "localhost")

	subscribePubsub  = getEnvVar("SUBSCRIBE_PUBSUB_NAME", "pubsub")
	subscribeTopic   = getEnvVar("SUBSCRIBE_TOPIC_NAME", "req-service")
	publishPubsub    = getEnvVar("PUBLISH_PUBSUB_NAME", "pubsub")
	publishPubsubTTL = getEnvVar("PUBLISH_PUBSUB_TTL", "60")

	targetRoot    = getEnvVar("TARGET_ROOT", "https://httpbin.org")
	targetVersion = getEnvVar("TARGET_VERSION", "v1.0.0")

	invokeTimeout   = getEnvVar("INVOKE_TIMEOUT", "60")
	publishTimeout  = getEnvVar("PUBLISH_TIMEOUT", "5")
	callbackTimeout = getEnvVar("CALLBACK_TIMEOUT", "5")
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	app := fiber.New(fiber.Config{
		ServerHeader:             version,
		DisableHeaderNormalizing: true,
		DisableStartupMessage:    true,
	})
	app.Use(helmet.New())

	cst := caster.New(nil)

	if debug {
		// /debug/pprof
		app.Use(pprof.New())
		app.Get("/dashboard", monitor.New())
	}

	// https://v1-rc3.docs.dapr.io/developing-applications/building-blocks/pubsub/howto-publish-subscribe/
	app.Get("/dapr/subscribe", func(c *fiber.Ctx) error {
		sub := []struct {
			Pubsubname string `json:"pubsubname"`
			Topic      string `json:"topic"`
			Route      string `json:"route"`
		}{{
			Pubsubname: subscribePubsub,
			Topic:      subscribeTopic,
			Route:      "/invoke",
		}}
		log.Info().Interface("sub", sub).Send()
		return c.JSON(sub)
	})

	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	app.All("/publish/:target/*", publishHandler(cst))
	app.Post("/log", logHandler)
	app.Post("/callback/:id", callbackHandler(cst))
	app.Post("/invoke", invokeHandler)

	go func() {
		log.Debug().Str("Server start", version).Send()
		if err := app.Listen(":" + serviceAddress); err != nil {
			log.Panic().Err(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	_ = <-c
	_ = app.Shutdown()
}

func logHandler(c *fiber.Ctx) error {
	b := map[string]interface{}{}

	if err := json.Unmarshal(c.Body(), &b); err != nil {
		return fiber.NewError(500, "CloudEvent Data Unmarchal failed.")
	}
	log.Info().
		Str("service", subscribeTopic).
		Str("version", targetVersion).
		Str("route", c.OriginalURL()).
		Interface("request", b).
		Send()
	return c.SendStatus(204)
}

// publishHandler start
func publishHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		before := time.Now()
		// set headers
		hd := map[string]string{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			hd[string(key)] = string(value)
		})

		r := reuestedData{
			Method:  c.Method(),
			URL:     c.BaseURL() + c.OriginalURL(),
			Headers: hd,
			Body:    c.Body(),
		}

		pub := publishData{
			Order:    &r,
			Callback: "http://" + myIP + serviceAddress,
		}

		ce := newCustomEvent(&pub, getTrace(c), c.Params("target"))
		log.Debug().
			Str("traceid", ce.TraceID).
			Str("service", subscribeTopic).
			Str("route", c.OriginalURL()).
			Interface("request", ce).
			Send()

		// https://v1-rc3.docs.dapr.io/reference/api/pubsub_api/#http-response-codes
		if err := publishrequestevent(ce); err != nil {
			return fiber.NewError(500, "Fail to publish event")
		}

		ch, ok := cst.Sub(nil, 1)
		if !ok {
			// https://docs.gofiber.io/api/fiber#newerror
			return fiber.NewError(782, "Caster subscription failed")
		}
		defer cst.Unsub(ch)

		rm := message{}

		for m := range ch {
			t := m.(message).ID
			if ce.ID.String() == t {
				rm = m.(message)
				break
			}
		}
		for k, v := range rm.Headers {
			c.Set(k, v)
		}
		st, _ := strconv.Atoi(rm.Status)

		after := time.Now()
		log.Info().
			Str("traceid", ce.TraceID).
			Str("service", subscribeTopic).
			Str("version", targetVersion).
			Str("route", c.OriginalURL()).
			Str("latency", after.Sub(before).String()).
			Interface("request", ce).
			Interface("response", rm).
			Send()
		return c.Status(st).Send(rm.Body)
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

func publishrequestevent(ce *customEvent) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	pubURL := "http://localhost:3500/v1.0/publish/" + publishPubsub + "/" + ce.Topic + "?metadata.ttlInSeconds=" + publishPubsubTTL
	req.SetRequestURI(pubURL)

	req.Header.SetMethod("POST")

	req.Header.SetContentType("application/cloudevents+json")
	req.Header.Set("traceparent", ce.TraceID)
	body, _ := json.Marshal(ce)

	req.SetBody(body)

	to, _ := strconv.Atoi(publishTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	fasthttp.AcquireResponse()
	return nil
}

func newCustomEvent(pub *publishData, tid string, targetTopic string) *customEvent {
	tz, _ := time.LoadLocation("UTC")
	return &customEvent{
		ID:     uuid.New(),
		Source: "nydus",
		// Type:            "com.dapr.event.sent",
		// SpecVersion:     "1.0",
		DataContentType: "application/json",
		Data:            pub,
		Topic:           targetTopic,
		TraceID:         tid,
		Time:            time.Now().In(tz).Format(time.RFC3339),
	}
}

type customEvent struct {
	ID     uuid.UUID `json:"id"`
	Source string    `json:"source"`
	// Type            string       `json:"type"`
	// SpecVersion     string       `json:"specversion"`
	DataContentType string       `json:"datacontenttype"`
	Topic           string       `json:"topic"`
	TraceID         string       `json:"traceid,omitempty"`
	Data            *publishData `json:"data"`
	Time            string       `json:"time"`
}

func (c *customEvent) propTrace() {
	c.Data.Order.Headers["traceparent"] = c.TraceID
}

type publishData struct {
	Order    *reuestedData          `json:"order"`
	Callback string                 `json:"callback"`
	Meta     map[string]interface{} `json:"meta"`
}

type reuestedData struct {
	Method  string             `json:"method"`
	URL     string             `json:"url"`
	Headers map[string]string  `json:"headers"`
	Body    jsonstd.RawMessage `json:"body"`
}

// callbackHandler start
func callbackHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		hd := map[string]string{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			hd[string(key)] = string(value)
		})

		m := message{
			ID:      c.Params("id"),
			Status:  c.Get("status"),
			Headers: hd,
			Body:    c.Body(),
		}

		if ok := cst.TryPub(m); !ok {
			return fiber.NewError(500, "Caster delivery failed")
		}
		return c.SendStatus(204)
	}
}

type message struct {
	ID      string
	Status  string
	Headers map[string]string
	Body    jsonstd.RawMessage
}

// invokeHandler start
func invokeHandler(c *fiber.Ctx) error {
	before := time.Now()
	ce := customEvent{}

	if err := json.Unmarshal(c.Body(), &ce); err != nil {
		return fiber.NewError(500, "CloudEvent Data Unmarchal failed.")
	}
	log.Debug().
		Str("traceid", ce.TraceID).
		Str("service", subscribeTopic).
		Str("route", "/invoke").
		Interface("request", ce).
		Send()

	ce.Data.updateHost(targetRoot)
	ce.propTrace()

	out, err := requesttoTarget(ce.Data.Order)
	if err != nil {
		out = &responseData{
			Body: []byte(err.Error()),
		}
	}

	cb := callback{
		Callback: ce.Data.Callback,
		ID:       ce.ID,
		Response: out,
	}

	if err := callbacktoSource(&cb); err != nil {
		return err
	}
	after := time.Now()

	log.Info().
		Str("traceid", ce.TraceID).
		Str("service", subscribeTopic).
		Str("version", targetVersion).
		Str("route", "/invoke").
		Str("latency", after.Sub(before).String()).
		Interface("request", ce).
		Interface("response", cb).
		Send()
	return c.JSON(fiber.Map{"success": true})
}

func requesttoTarget(in *reuestedData) (out *responseData, err error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(in.URL)
	req.Header.SetMethod(strings.ToUpper(in.Method))
	for k, v := range in.Headers {
		req.Header.Set(k, v)
	}

	if string(in.Body) != "null" {
		req.SetBody([]byte(in.Body))
	}

	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	err = fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return nil, err
	}

	outraw := fasthttp.AcquireResponse()
	resp.CopyTo(outraw)

	hd := map[string]string{}
	outraw.Header.VisitAll(func(key, value []byte) {
		hd[string(key)] = string(value)
	})

	out = &responseData{
		Status:  resp.StatusCode(),
		Headers: hd,
		Body:    outraw.Body(),
	}

	return out, nil
}

func callbacktoSource(cb *callback) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(cb.Callback + "/callback/" + cb.ID.String())
	req.Header.SetMethod(strings.ToUpper("POST"))
	req.Header.Set("status", strconv.Itoa(cb.Response.Status))

	for k, v := range cb.Response.Headers {
		req.Header.Set(k, v)
	}
	req.SetBody([]byte(cb.Response.Body))

	to, _ := strconv.Atoi(callbackTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	out := fasthttp.AcquireResponse()
	resp.CopyTo(out)
	return nil
}

func (p *publishData) updateHost(r string) {
	t, _ := url.Parse(p.Order.URL)
	p.Order.URL = setHost(r, t)
}

type callback struct {
	Callback string
	ID       uuid.UUID
	Response *responseData
}

type responseData struct {
	Status  int                `json:"status"`
	Headers map[string]string  `json:"headers"`
	Body    jsonstd.RawMessage `json:"body"`
}

func setHost(r string, u *url.URL) string {
	t, _ := url.Parse(r)
	u.Scheme = t.Scheme
	u.Host = t.Host + t.Path
	u.Path = strings.ReplaceAll(u.Path, "/publish/"+subscribeTopic, "")
	h, _ := url.PathUnescape(u.String())
	return h
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
