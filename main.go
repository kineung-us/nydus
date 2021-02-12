package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/guiguan/caster"
	"github.com/valyala/fasthttp"
)

var (
	serviceAddress = getEnvVar("ADDRESS", ":5000")
	myIP           = getEnvVar("MY_POD_IP", "")

	sourcePubSub = getEnvVar("SOURCE_PUBSUB_NAME", "pubsub")
	sourceTopic  = getEnvVar("SOURCE_TOPIC_NAME", "req-service")
	pubsubTTL    = getEnvVar("PUBSUB_TTL", "60")
	pubURL       = "http://localhost:3500/v1.0/publish/" + sourcePubSub + "/" + sourceTopic
	headersList  = getEnvVar("PROPAGATE_HEADER_LIST", "dialog-session-id, dialog-transaction-id")

	targetRoot = getEnvVar("TARGET_ROOT", "http://localhost:3000")

	invokeTimeout   = getEnvVar("INVOKE_TIMEOUT", "60")
	publishTIimeout = getEnvVar("PUBLISH_TIMEOUT", "60")
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

	app.Use(logger.New())

	cst := caster.New(nil)

	// /debug/pprof
	// app.Use(pprof.New())
	// app.Get("/dashboard", monitor.New())

	// https://v1-rc3.docs.dapr.io/developing-applications/building-blocks/pubsub/howto-publish-subscribe/
	app.Get("/dapr/subscribe", func(c *fiber.Ctx) error {
		sub := []subscribe{}
		sub = append(sub, subscribe{
			Pubsubname: sourcePubSub,
			Topic:      sourceTopic,
			Route:      "/invoke",
		})
		return c.JSON(sub)
	})

	app.Post("/publish/:target/*", publish(cst))
	app.Post("/callback/:id", callback(cst))
	app.Post("/invoke", invoke)

	go func() {
		if err := app.Listen(serviceAddress); err != nil {
			fmt.Errorf("error to start")
		}
	}()

	c := make(chan os.Signal, 1)                                    // 1Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // When an interrupt or termination signal is sent, notify the channel

	_ = <-c // This blocks the main thread until an interrupt is received
	_ = app.Shutdown()
}

// 요청을 받으면
// 요청 처리 쪽에 퍼블리쉬
// 		이게 post 요청인데 publish body로 해야 함.
// 응답 기다림
// caster로 오면 응답 전달
// 종료
func publish(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		fmt.Println(c.Params("*"))

		if headersList != "" {
			c.Get("Content-Type")
		}

		r := dataRequested{
			Method:  c.Method(),
			URL:     c.BaseURL() + c.OriginalURL(),
			Headers: c.Request().Header.Header(),
			Body:    c.Body(),
		}

		pub := publishData{
			Order:    &r,
			Callback: myIP,
		}

		ce := newCustomEvent(&pub, c.Params("target"))

		publishrequestevent(ce)

		ch, ok := cst.Sub(nil, 1)
		if !ok {
			// https://docs.gofiber.io/api/fiber#newerror
			return fiber.NewError(782, "Custom error message")
		}
		defer cst.Unsub(ch)

		body := []byte{}

		fmt.Println(ce.ID.String())

		for m := range ch {
			t := m.(message).ID
			if ce.ID.String() == t {
				body = m.(message).Body
				break
			}
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

	// for k, v := range r.Headers {
	// req.Header.Add(k, v)
	// }

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

	// just for demo
	fasthttp.AcquireResponse()
	return nil
}

// 요청이 들어오면
// go 루틴 caster로 id 기준 전달
// 받았음 리턴
func callback(cst *caster.Caster) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ok := cst.TryPub(message{
			ID:   c.Params("id"),
			Body: c.Body(),
		})
		if !ok {
			return fiber.NewError(782, "Custom error message")
		}
		return c.SendString("send!")
	}
}

// subscribe 주소임
// 요청 바디를 받아서
// go 루틴 시작 (아래로)
// 받았다는 응답 하고 마무리
// go 루틴 시작하는데 이제부터 아래
// root 주소로 주소 조립후
// post 수행
// 기다림
// 응답을 받아서
// hostIP 에게 post로 전달
// 그리고 종료
func invoke(c *fiber.Ctx) error {
	body := c.Body()
	fmt.Println(string(body))
	// req를 post로 url에 전달
	// targetRoot + path 등등 수행

	// Requesttotarget()
	// Callbacktosource()

	return c.JSON(fiber.Map{"success": true})
}

type message struct {
	ID   string
	Body []byte
}

type subscribe struct {
	Pubsubname string `json:"pubsubname"`
	Topic      string `json:"topic"`
	Route      string `json:"route"`
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}

func newCustomEvent(pub *publishData, targetTopic string) *customEvent {
	return &customEvent{
		ID:              uuid.New(),
		Source:          "nydus",
		Type:            "com.dapr.event.sent",
		SpecVersion:     "1.0",
		DataContentType: "application/json",
		Data:            pub,
		Topic:           targetTopic,
		PubsubName:      sourcePubSub,
		Time:            time.Now().Format(time.RFC3339),
	}
}

type customEvent struct {
	ID              uuid.UUID    `json:"id"`
	Source          string       `json:"source"`
	Type            string       `json:"type"`
	SpecVersion     string       `json:"specversion"`
	DataContentType string       `json:"datacontenttype"`
	Data            *publishData `json:"data"`
	Topic           string       `json:"topic"`
	PubsubName      string       `json:"pubsubname"`
	Time            string       `json:"Time"`
}

type publishData struct {
	Order    *dataRequested         `json:"order"`
	Callback string                 `json:"callback"`
	Meta     map[string]interface{} `json:"meta"`
}

type dataRequested struct {
	Method  string `json:"method"`
	URL     string `json:"url"`
	Headers []byte `json:"headers"`
	Body    []byte `json:"body"`
}

// FastPostByte  do  POST request via fasthttp
func FastPostByte(uri string, r *dataRequested) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(r.URL)

	// for k, v := range r.Headers {
	// req.Header.Add(k, v)
	// }

	req.Header.SetMethod(strings.ToUpper(r.Method))

	req.Header.SetContentType("application/json")
	req.SetBody(r.Body)

	to, _ := strconv.Atoi(invokeTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {

		return nil, err
	}

	// just for demo
	out := fasthttp.AcquireResponse()
	resp.CopyTo(out)

	return out, nil
}

func setHost(r string, u *url.URL) string {
	t, _ := url.Parse(r)
	u.Scheme = t.Scheme
	u.Host = t.Host + t.Path
	h, _ := url.PathUnescape(u.String())
	return h
}

func getHeaderList(hl string) map[string]string {
	return map[string]string{}
}

func publishOrder(pd *publishData) (string, error) {
	// application/cloudevents+json
	return "", nil
}
