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
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	version = "nydus-v0.0.1"

	debug, _ = strconv.ParseBool(getEnvVar("DEBUG", "false"))

	serviceAddress = getEnvVar("ADDRESS", ":5000")
	myIP           = getEnvVar("MY_POD_IP", "localhost")

	sourcePubSub = getEnvVar("SOURCE_PUBSUB_NAME", "pubsub")
	sourceTopic  = getEnvVar("SOURCE_TOPIC_NAME", "req-service")
	pubsubTTL    = getEnvVar("PUBSUB_TTL", "60")
	pubURL       = "http://localhost:3500/v1.0/publish/" + sourcePubSub + "/" + sourceTopic + "?metadata.ttlInSeconds=" + pubsubTTL

	targetRoot = getEnvVar("TARGET_ROOT", "http://localhost:3000")

	pheader1 = getEnvVar("PROPAGATE_HEADER_1", "dsid=dialog-session-id")
	pheader2 = getEnvVar("PROPAGATE_HEADER_2", "dtid=dialog-transaction-id")
	pheader3 = getEnvVar("PROPAGATE_HEADER_3", "")

	invokeTimeout   = getEnvVar("INVOKE_TIMEOUT", "60")
	publishTimeout  = getEnvVar("PUBLISH_TIMEOUT", "5")
	callbackTimeout = getEnvVar("CALLBACK_TIMEOUT", "5")
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339

	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  version,
	})

	app.Use(helmet.New())
	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			log.Print(string(c.Context().Path()))
			return false
		},
		Format:       logFormatwithHeaders(),
		TimeFormat:   time.RFC3339,
		TimeZone:     "UTC",
		TimeInterval: 0,
		Output:       os.Stderr,
	}))

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
			Pubsubname: sourcePubSub,
			Topic:      sourceTopic,
			Route:      "/invoke",
		}}
		return c.JSON(sub)
	})

	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	app.All("/publish/:target/*", publishHandler(cst))
	app.Post("/callback/:id", callbackHandler(cst))
	app.Post("/invoke", invokeHandler)

	go func() {
		if err := app.Listen(serviceAddress); err != nil {
			log.Panic().Err(err)
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
			Callback: "http://" + myIP + serviceAddress,
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
	log.Print(string(body))
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
			Status:  c.Get("status"),
			Headers: hd,
			Body:    c.Body(),
		})
		if !ok {
			return fiber.NewError(500, "Caster delivery failed")
		}
		return c.Status(204).JSON(fiber.Map{"success": true})
	}
}

type message struct {
	ID      string
	Status  string
	Headers map[string]string
	Body    []byte
}

// invokeHandler start
func invokeHandler(c *fiber.Ctx) error {
	ce := customEvent{}

	if err := json.Unmarshal(c.Body(), &ce); err != nil {
		return fiber.NewError(500, "CloudEvent Data Unmarchal failed.")
	}

	ce.Data.updateHost(targetRoot)

	out, err := requesttoTarget(ce.Data.Order)
	if err != nil {
		out = &responseData{
			Body: []byte(err.Error()),
		}
	}

	out.Headers = propHeader(ce.Data.Order.Headers, out.Headers)
	cb := callback{
		Callback: ce.Data.Callback,
		ID:       ce.ID,
		Response: out,
	}
	if err := callbacktoSource(&cb); err != nil {
		return err
	}
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
		req.Header.Add(k, v)
	}

	if string(in.Body) != "null" {
		req.SetBody([]byte(in.Body))
	}

	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	before := time.Now()
	err = fasthttp.DoTimeout(req, resp, timeOut)
	after := time.Now()
	if err != nil {
		return nil, err
	}

	outraw := fasthttp.AcquireResponse()
	resp.CopyTo(outraw)

	hd := map[string]string{}
	outraw.Header.VisitAll(func(key, value []byte) {
		hd[string(key)] = string(value)
	})
	hd["invoke-latency"] = after.Sub(before).String()

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

	log.Print(cb.Callback + "/callback/" + cb.ID.String())

	for k, v := range cb.Response.Headers {
		req.Header.Add(k, v)
	}
	req.SetBody([]byte(cb.Response.Body))

	to, _ := strconv.Atoi(callbackTimeout)
	timeOut := time.Duration(to) * time.Second

	log.Print(req.URI().String())
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

func logFormatwithHeaders() string {
	logFormatStart := "{\"time\": \"${time}\", \"route\": \"${route}\", \"status\": ${status}, \"latency\": \"${latency}\", \"invoke-latency\": \"${header:invoke-latency}\", "
	logFormatPheaders := ""
	logFormatEnd := "\"body\": ${body}, \"resBody\": ${resBody}}\n"

	if pheader1 != "" {
		logFormatPheaders = propHeadertoLog(pheader1, logFormatPheaders)
	}
	if pheader2 != "" {
		logFormatPheaders = propHeadertoLog(pheader2, logFormatPheaders)
	}
	if pheader3 != "" {
		logFormatPheaders = propHeadertoLog(pheader3, logFormatPheaders)
	}
	return logFormatStart + logFormatPheaders + logFormatEnd
}

func propHeadertoLog(pheader string, log string) string {
	head := strings.Split(pheader, "=")
	log = log + `"` + head[0] + `": "${header:` + head[1] + `}", `
	return log
}

func propHeader(in map[string]string, out map[string]string) map[string]string {
	for k, v := range in {
		pks := []string{}
		if pheader1 != "" {
			tem := strings.Split(pheader1, "=")
			pks = append(pks, tem[1])
		}
		if pheader2 != "" {
			tem := strings.Split(pheader2, "=")
			pks = append(pks, tem[1])
		}
		if pheader3 != "" {
			tem := strings.Split(pheader3, "=")
			pks = append(pks, tem[1])
		}
		for _, pk := range pks {
			if strings.Contains(strings.ToLower(k), strings.ToLower(pk)) {
				out[k] = v
			}
		}
	}
	return out
}
