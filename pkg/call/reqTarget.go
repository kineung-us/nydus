package call

import (
	"strconv"
	"strings"
	"time"

	"nydus/pkg/body"

	"github.com/valyala/fasthttp"
)

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

	b, errm := body.Marshal(in.Body, in.Headers["Content-Type"])
	if errm != nil {
		return nil, errm
	}

	if b != nil {
		req.SetBody(b)
	}

	to, _ := strconv.Atoi(ivkTimeout)
	timeOut := time.Duration(to) * time.Second

	if err := fasthttp.DoTimeout(req, resp, timeOut); err != nil {
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
