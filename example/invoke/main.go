package main

import (
	"encoding/json"
	jsonstd "encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var (
	sourceTopic = "hermesapi"
	targetRoot  = "http://localhost:8080"
)

func main() {
	ce := customEvent{}

	raw := []byte(``)

	json.Unmarshal(raw, &ce)

	ce.Data.updateHost(targetRoot)

	out, err := requesttoTarget(ce.Data.Order)
	if err != nil {
		out = &responseData{
			Body: []byte(err.Error()),
		}
	}
	log.Print(string(out.Body))
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

	log.Print("request data!")
	log.Print(string(req.Header.Header()))
	log.Print(req.URI())
	log.Print(req.String())
	log.Print(string(req.Body()))

	to, _ := strconv.Atoi("60")
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

func newCustomEvent(pub *publishData, targetTopic string) *customEvent {
	tz, _ := time.LoadLocation("UTC")
	return &customEvent{
		ID:     uuid.New(),
		Source: "nydus",
		// Type:            "com.dapr.event.sent",
		// SpecVersion:     "1.0",
		DataContentType: "application/json",
		Data:            pub,
		Topic:           targetTopic,
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
	Topic           string       `json:"topic"`
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

func (p *publishData) updateHost(r string) {
	t, _ := url.Parse(p.Order.URL)
	p.Order.URL = setHost(r, t)
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
