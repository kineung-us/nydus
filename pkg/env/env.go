package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	Version  = "dev"
	Nversion = "nydus-" + Version
	Debug, _ = strconv.ParseBool(getEnvVar("DEBUG", "false"))

	ServiceAddress = getEnvVar("NYDUS_HTTP_PORT", "5000")
	ServiceIP      = getEnvVar("NYDUS_HOST_IP", "localhost")

	SubscribePubsub = getEnvRequired("SUBSCRIBE_PUBSUB_NAME")
	SubscribeTopic  = getEnvRequired("SUBSCRIBE_TOPIC_NAME")

	PublishPubsub    = getEnvRequired("PUBLISH_PUBSUB_NAME")
	PublishPubsubTTL = getEnvVar("PUBLISH_PUBSUB_TTL", "120")

	TargetRoot        = getEnvRequired("TARGET_ROOT")
	TargetHealthzAddr = getEnvRequired("TARGET_HEALTHZ_ADDR")
	TargetVersion     = getEnvRequired("TARGET_VERSION")

	InvokeTimeout, _   = strconv.Atoi(getEnvVar("INVOKE_TIMEOUT", "100"))
	PublishTimeout, _  = strconv.Atoi(getEnvVar("PUBLISH_TIMEOUT", "10"))
	CallbackTimeout, _ = strconv.Atoi(getEnvVar("CALLBACK_TIMEOUT", "10"))

	DaprHealthzAddr       = getEnvVar("DAPR_HEALTHZ_ADDR", "http://localhost:3500/v1.0/healthz")
	DaprHealthzTimeout, _ = strconv.Atoi(getEnvVar("DAPR_HEALTHZ_TIMEOUT", "5"))

	ClientMaxConnsPerHost, _   = strconv.Atoi(getEnvVar("CLIENT_MAX_CONNS_PER_HOST", "10000"))
	ClientReadTimeoutSec, _    = strconv.Atoi(getEnvVar("CLIENT_READ_TIMEOUT", "100"))
	ClientWriteTimeoutSec, _   = strconv.Atoi(getEnvVar("CLIENT_WRITE_TIMEOUT", "10"))
	ClientHeaderNormalizing, _ = strconv.ParseBool(getEnvVar("CLIENT_HEADER_NORMALIZING", "false"))

	ServerReadTimeoutSec, _    = strconv.Atoi(getEnvVar("SERVER_READ_TIMEOUT", "100"))
	ServerWriteTimeoutSec, _   = strconv.Atoi(getEnvVar("SERVER_WRITE_TIMEOUT", "10"))
	ServerIdleTimeoutSec, _    = strconv.Atoi(getEnvVar("SERVER_IDLE_TIMEOUT", "100"))
	ServerHeaderNormalizing, _ = strconv.ParseBool(getEnvVar("SERVER_HEADER_NORMALIZING", "false"))

	XMLtoString, _ = strconv.ParseBool(getEnvVar("XML_TO_STRING", "true"))
	DFTtoString, _ = strconv.ParseBool(getEnvVar("DEFAULT_TO_STRING", "false"))
)

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		if val != "" {
			return strings.TrimSpace(val)
		}
	}
	return fallbackValue
}

func getEnvRequired(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("environment variable(key: \"%s\") is required", key))
	}
	return val
}
