package call

import (
	"nydus/pkg/env"
	"path"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ppubsub     = env.PublishPubsub
	ttl         = env.PublishPubsubTTL
	pubTimeout  = env.PublishTimeout
	ivkTimeout  = env.InvokeTimeout
	cbTimeout   = env.CallbackTimeout
	hthzTimeout = time.Second * time.Duration(env.HealthzTimeout)

	dhzaddr    = env.DaprHealthzAddr
	dhzTimeout = env.DaprHealthzTimeout

	troot   = env.TargetRoot
	thzpath = env.TargetHealthzPath
	thzaddr = ""

	clHeaderNorm = env.ClientHeaderNormalizing

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               env.ClientMaxConnsPerHost,
		ReadTimeout:                   time.Second * time.Duration(env.ClientReadTimeoutSec),
		WriteTimeout:                  time.Second * time.Duration(env.ClientWriteTimeoutSec),
		DisableHeaderNamesNormalizing: !clHeaderNorm,
	}
)

func init() {
	troot.Path = path.Join(troot.Path, thzpath)
	thzaddr = troot.String()
}

func DaprHealthChk() bool {
	log.Debug().Str("func", "DaprHealthChk").Send()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(dhzaddr)
	if err := client.DoTimeout(req, resp, time.Duration(dhzTimeout)*time.Second); err != nil {
		return false
	}

	if resp.StatusCode() == 204 {
		return true
	}
	return false
}

func TargetHealthChk() int {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(thzaddr)

	if err := client.DoTimeout(req, resp, hthzTimeout); err != nil {
		return 400
	}

	log.Debug().Str("func", "TargetHealthChk").
		Int("return", resp.StatusCode()).Send()

	return resp.StatusCode()
}
