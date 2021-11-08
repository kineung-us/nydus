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

	ppubsub      = env.PublishPubsub
	ttl          = env.PublishPubsubTTL
	pubTimeout   = env.PublishTimeout
	ivkTimeout   = env.InvokeTimeout
	cbTimeout    = env.CallbackTimeout
	troot        = env.TargetRoot
	thzaddr      = env.TargetHealthzAddr
	dhzaddr      = env.DaprHealthzAddr
	dhzTimeout   = time.Duration(env.DaprHealthzTimeout) * time.Second
	clHeaderNorm = env.ClientHeaderNormalizing

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               env.ClientMaxConnsPerHost,
		ReadTimeout:                   time.Second * time.Duration(env.ClientReadTimeoutSec),
		WriteTimeout:                  time.Second * time.Duration(env.ClientWriteTimeoutSec),
		DisableHeaderNamesNormalizing: !clHeaderNorm,
	}
)

func DaprHealthChk() bool {
	log.Debug().Str("func", "DaprHealthChk").Str("address", dhzaddr).Send()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(dhzaddr)
	if err := client.DoTimeout(req, resp, dhzTimeout); err != nil {
		log.Error().Stack().Err(err).Send()
		return false
	}

	if resp.StatusCode() == 204 {
		return true
	}
	return false
}

func TargetHealthChk() int {
	t := path.Join(troot, thzaddr)
	log.Debug().Str("func", "TargetHealthChk").
		Str("address", t).Send()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(t)

	if err := client.DoTimeout(req, resp, dhzTimeout); err != nil {
		log.Error().Stack().Err(err).Send()
		return 400
	}

	return resp.StatusCode()
}
