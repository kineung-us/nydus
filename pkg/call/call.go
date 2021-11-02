package call

import (
	"nydus/pkg/env"
	"path"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ppubsub    = env.PublishPubsub
	ttl        = env.PublishPubsubTTL
	pubTimeout = env.PublishTimeout
	ivkTimeout = env.InvokeTimeout
	cbTimeout  = env.CallbackTimeout
	troot      = env.TargetRoot
	thzaddr    = env.TargetHealthzAddr
	dhzaddr    = env.DaprHealthzAddr
	dhzTimeout = env.DaprHealthzTimeout

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               10000,
		ReadTimeout:                   time.Minute,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true,
		MaxConnWaitTimeout:            time.Second,
	}
)

func DaprHealthChk() bool {
	log.Debug().Str("func", "DaprHealthChk").Send()
	chk := false
	to, _ := strconv.Atoi(dhzTimeout)
	timeOut := time.Duration(to) * time.Second
	st, _, _ := fasthttp.GetTimeout(nil, dhzaddr, timeOut)
	if st == 204 {
		chk = true
	}
	return chk
}

func TargetHealthChk() int {
	log.Debug().Str("func", "TargetHealthChk").Send()
	to, _ := strconv.Atoi(dhzTimeout)
	timeOut := time.Duration(to) * time.Second
	st, _, _ := fasthttp.GetTimeout(nil, path.Join(troot, thzaddr), timeOut)
	return st
}
