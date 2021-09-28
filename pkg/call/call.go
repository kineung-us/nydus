package call

import (
	"nydus/pkg/env"
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
	dhzaddr    = env.DaprHealthzAddr
	dhzTimeout = env.DaprHealthzTimeout
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
