package call

import (
	"strings"
	"time"

	"nydus/pkg/body"

	"github.com/valyala/fasthttp"
)

func RequesttoTarget(in *body.RequestedData) (out *body.ResponseData, err error) {
	req := fasthttp.AcquireRequest()
	if !clHeaderNorm {
		req.Header.DisableNormalizing()
	}
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

	b, errm := body.Marshal(in.Body, in.Headers["Content-Type"])
	if errm != nil {
		return nil, errm
	}

	if b != nil {
		req.SetBody(b)
	}

	timeOut := time.Duration(ivkTimeout) * time.Second

	if err := client.DoTimeout(req, resp, timeOut); err != nil {
		return nil, err
	}

	hd := map[string]string{}
	resp.Header.VisitAll(func(key, value []byte) {
		hd[string(key)] = string(value)
	})

	out = &body.ResponseData{
		Status:  resp.StatusCode(),
		Headers: hd,
		Body:    resp.Body(),
	}

	return out, nil
}
