package main

import (
	jsonstd "encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/guiguan/caster"
	"github.com/valyala/fasthttp"
)

var (
	json           = jsoniter.ConfigCompatibleWithStandardLibrary
	serviceAddress = getEnvVar("ADDRESS", ":5000")
	myIP           = getEnvVar("MY_POD_IP", "http://localhost")

	sourcePubSub = getEnvVar("SOURCE_PUBSUB_NAME", "pubsub")
	sourceTopic  = getEnvVar("SOURCE_TOPIC_NAME", "req-service")
	pubsubTTL    = getEnvVar("PUBSUB_TTL", "60")
	pubURL       = "http://localhost:3500/v1.0/publish/" + sourcePubSub + "/" + sourceTopic

	// TODO: 설계가 필요함. tid를 meta로 옮길지 등등이 필요함.
	headersList = getEnvVar("PROPAGATE_HEADER_LIST", "dialog-session-id, dialog-transaction-id")

	targetRoot = getEnvVar("TARGET_ROOT", "http://localhost:3000")

	invokeTimeout   = getEnvVar("INVOKE_TIMEOUT", "60")
	publishTimeout  = getEnvVar("PUBLISH_TIMEOUT", "5")
	callbackTimeout = getEnvVar("CALLBACK_TIMEOUT", "5")
)

func main() {
	version := "nydus-v0.0.1"

	// if myIP == "" {
	// 	panic("MY_POD_IP is required.")
	// }

	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  version,
	})

	app.Use(logger.New(logger.Config{
		Format:     "{time: ${time}, route: ${route}, status: ${status}, latency: ${latency}, body: ${body}, resBody: ${resBody}}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
	}))

	cst := caster.New(nil)

	// /debug/pprof
	// app.Use(pprof.New())
	// app.Get("/dashboard", monitor.New())

	// https://v1-rc3.docs.dapr.io/developing-applications/building-blocks/pubsub/howto-publish-subscribe/
	app.Get("/dapr/subscribe", func(c *fiber.Ctx) error {
		sub := []struct {
			Pubsubname string `json:"pubsubname"`
			Topic      string `json:"topic"`
			Route      string `json:"route"`
		}{{
			Pubsubname: sourcePubSub,
			Topic:      sourceTopic,
			Route:      "/invoke",
		}}
		return c.JSON(sub)
	})

	app.Post("/publish/:target/*", publishHandler(cst))
	app.Post("/callback/:id", callbackHandler(cst))
	app.Post("/invoke", invokeHandler)

	go func() {
		if err := app.Listen(serviceAddress); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	_ = <-c
	_ = app.Shutdown()
}

// publishHandler start
func publishHandler(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

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
			Callback: myIP + serviceAddress,
		}

		ce := newCustomEvent(&pub, c.Params("target"))

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

		body := []byte{}
		headers := map[string]string{}

		for m := range ch {
			t := m.(message).ID
			if ce.ID.String() == t {
				body = m.(message).Body
				headers = m.(message).Headers
				break
			}
		}
		for k, v := range headers {
			c.Set(k, v)
		}
		return c.Send(body)
	}
}

func publishrequestevent(ce *customEvent) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(pubURL)

	req.Header.SetMethod("POST")

	req.Header.SetContentType("application/cloudevents+json")
	body, _ := json.Marshal(ce)
	req.SetBody(body)

	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	fasthttp.AcquireResponse()
	return nil
}

func newCustomEvent(pub *publishData, targetTopic string) *customEvent {
	tz, _ := time.LoadLocation("UTC")
	return &customEvent{
		ID:     uuid.New(),
		Source: "nydus",
		// Type:            "com.dapr.event.sent",
		// SpecVersion:     "1.0",
		DataContentType: "application/json",
		Data:            pub,
		// Topic:           targetTopic,
		// PubsubName:      sourcePubSub,
		Time: time.Now().In(tz).Format(time.RFC3339),
	}
}

type customEvent struct {
	ID     uuid.UUID `json:"id"`
	Source string    `json:"source"`
	// Type            string       `json:"type"`
	// SpecVersion     string       `json:"specversion"`
	DataContentType string       `json:"datacontenttype"`
	Data            *publishData `json:"data"`
	Time            string       `json:"time"`
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
		ok := cst.TryPub(message{
			ID:      c.Params("id"),
			Headers: hd,
			Body:    c.Body(),
		})
		if !ok {
			return fiber.NewError(500, "Caster delivery failed")
		}
		return c.SendStatus(204)
	}
}

type message struct {
	ID      string
	Headers map[string]string
	Body    []byte
}

// invokeHandler start
func invokeHandler(c *fiber.Ctx) error {
	ce := customEvent{}

	if err := json.Unmarshal(c.Body(), &ce); err != nil {
		fmt.Println("error")
		log.Fatalln(err)
	}

	ce.Data.updateHost(targetRoot)
	fmt.Println("invoke")
	fmt.Println(ce.Data.Order.URL)

	out, err := requesttoTarget(ce.Data.Order)
	if err != nil {
		out = &reuestedData{
			Body: []byte(err.Error()),
		}
	}

	ce.Data.Order = out

	if err := callbacktoSource(&ce); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func requesttoTarget(in *reuestedData) (out *reuestedData, err error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(in.URL)
	req.Header.SetMethod(strings.ToUpper(in.Method))
	for k, v := range in.Headers {
		req.Header.Add(k, v)
	}
	req.SetBody([]byte(in.Body))
	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	err = fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return nil, err
	}

	// just for demo
	outraw := fasthttp.AcquireResponse()
	resp.CopyTo(outraw)

	hd := map[string]string{}
	outraw.Header.VisitAll(func(key, value []byte) {
		hd[string(key)] = string(value)
	})

	out = &reuestedData{
		Headers: hd,
		Body:    outraw.Body(),
	}

	return out, nil
}

func callbacktoSource(ce *customEvent) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(ce.Data.Callback + "/callback/" + ce.ID.String())
	req.Header.SetMethod(strings.ToUpper("POST"))
	req.Header.SetContentType("application/json")
	for k, v := range ce.Data.Order.Headers {
		req.Header.Add(k, v)
	}
	req.SetBody([]byte(ce.Data.Order.Body))

	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	// just for demo
	out := fasthttp.AcquireResponse()
	resp.CopyTo(out)
	return nil
}

func (p *publishData) updateHost(r string) {
	t, _ := url.Parse(p.Order.URL)
	p.Order.URL = setHost(r, t)
}

func setHost(r string, u *url.URL) string {
	t, _ := url.Parse(r)
	u.Scheme = t.Scheme
	u.Host = t.Host + t.Path
	u.Path = strings.ReplaceAll(u.Path, "/publish/"+sourceTopic, "")
	h, _ := url.PathUnescape(u.String())
	return h
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
