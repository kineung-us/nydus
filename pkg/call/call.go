package call

import (
	"strconv"
	"strings"
	"time"

	"nydus/pkg/body"
	"nydus/pkg/env"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ppubsub    = env.PublishPubsub
	ttl        = env.PublishPubsubTTL
	pubTimeout = env.PublishTimeout
	ivkTimeout = env.InvokeTimeout
	cbTimeout  = env.CallbackTimeout
)

func Publishrequestevent(ce *body.CustomEvent) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	pubURL := "http://localhost:3500/v1.0/publish/" + ppubsub + "/" + ce.Topic + "?metadata.ttlInSeconds=" + ttl
	req.SetRequestURI(pubURL)

	req.Header.SetMethod("POST")

	req.Header.SetContentType("application/cloudevents+json")
	req.Header.Set("traceparent", ce.TraceID)
	body, _ := json.Marshal(ce)

	req.SetBody(body)

	to, _ := strconv.Atoi(pubTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	fasthttp.AcquireResponse()
	return nil
}

func RequesttoTarget(in *body.RequestedData) (out *body.ResponseData, err error) {
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

	b, err := body.Marshal(in.Body, in.Headers["Content-Type"])
	if err != nil {
		return nil, err
	}

	if b != nil {
		req.SetBody(b)
	}

	to, _ := strconv.Atoi(ivkTimeout)
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

	out = &body.ResponseData{
		Status:  resp.StatusCode(),
		Headers: hd,
		Body:    outraw.Body(),
	}

	return out, nil
}

func CallbacktoSource(cb *body.Callback) error {
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

	req.SetBody(cb.Response.Body.([]byte))

	to, _ := strconv.Atoi(cbTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	out := fasthttp.AcquireResponse()
	resp.CopyTo(out)
	return nil
}
