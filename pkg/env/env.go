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
	PublishPubsubTTL = getEnvVar("PUBLISH_PUBSUB_TTL", "60")

	TargetRoot    = getEnvRequired("TARGET_ROOT")
	TargetVersion = getEnvRequired("TARGET_VERSION")

	InvokeTimeout   = getEnvVar("INVOKE_TIMEOUT", "60")
	PublishTimeout  = getEnvVar("PUBLISH_TIMEOUT", "5")
	CallbackTimeout = getEnvVar("CALLBACK_TIMEOUT", "5")

	DaprHealthzAddr    = getEnvVar("DAPR_HEALTHZ_ADDR", "http://localhost:3500/v1.0/healthz")
	DaprHealthzTimeout = getEnvVar("DAPR_HEALTHZ_TIMEOUT", "5")

	XMLtoString, _ = strconv.ParseBool(getEnvVar("XML_TO_STRING", "true"))
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
