package call

import (
	"time"

	"nydus/pkg/body"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func Publishrequestevent(ce *body.CustomEvent) error {
	log.Debug().
		Str("func", "Publishrequestevent").
		Send()

	req := fasthttp.AcquireRequest()
	if !clHeaderNorm {
		req.Header.DisableNormalizing()
	}
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

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("func", "Publishrequestevent").
		Str("pubURL", pubURL).
		Interface("request", ce).
		Send()

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("func", "Publishrequestevent").
		Interface("requestObj", req).
		Send()

	timeOut := time.Duration(pubTimeout) * time.Second

	if err := client.DoTimeout(req, resp, timeOut); err != nil {
		return err
	}

	log.Debug().
		Str("traceid", ce.TraceID).
		Str("func", "Publishrequestevent-publishend").
		Int("StateCode", resp.StatusCode()).
		Str("response", string(resp.Body())).
		Send()
	return nil
}
