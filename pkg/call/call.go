package call

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"nydus/pkg/body"
	"nydus/pkg/env"

	"github.com/valyala/fasthttp"
)

func Publishrequestevent(ce *body.CustomEvent) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	pubURL := "http://localhost:3500/v1.0/publish/" + env.PublishPubsub + "/" + ce.Topic + "?metadata.ttlInSeconds=" + env.PublishPubsubTTL
	req.SetRequestURI(pubURL)

	req.Header.SetMethod("POST")

	req.Header.SetContentType("application/cloudevents+json")
	req.Header.Set("traceparent", ce.TraceID)
	body, _ := json.Marshal(ce)

	req.SetBody(body)

	to, _ := strconv.Atoi(env.PublishTimeout)
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

	if in.Method != "GET" {
		req.SetBody(b)
	}

	to, _ := strconv.Atoi(env.InvokeTimeout)
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

	to, _ := strconv.Atoi(env.CallbackTimeout)
	timeOut := time.Duration(to) * time.Second

	err := fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return err
	}

	out := fasthttp.AcquireResponse()
	resp.CopyTo(out)
	return nil
}
