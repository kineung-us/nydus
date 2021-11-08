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
	dhzTimeout   = env.DaprHealthzTimeout
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
	log.Debug().Str("func", "DaprHealthChk").Send()
	chk := false
	st, _, _ := client.GetTimeout(nil, dhzaddr, time.Duration(dhzTimeout)*time.Second)
	if st == 204 {
		chk = true
	}
	return chk
}

func TargetHealthChk() int {
	log.Debug().Str("func", "TargetHealthChk").Send()
	log.Debug().Str("chk-path", path.Join(troot, thzaddr)).Send()
	st, _, _ := client.GetTimeout(nil, path.Join(troot, thzaddr), time.Duration(dhzTimeout)*time.Second)
	return st
}
