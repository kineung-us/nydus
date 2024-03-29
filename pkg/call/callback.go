package call

import (
	"strconv"
	"strings"
	"time"

	"nydus/pkg/body"

	"github.com/valyala/fasthttp"
)

func CallbacktoSource(cb *body.Callback) error {
	req := fasthttp.AcquireRequest()
	if !clHeaderNorm {
		req.Header.DisableNormalizing()
	}
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

	timeOut := time.Duration(cbTimeout) * time.Second

	if err := client.DoTimeout(req, resp, timeOut); err != nil {
		return err
	}
	return nil
}
