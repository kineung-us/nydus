package main

import (
	"context"
	"log"
	"os"
	"strings"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "", 0)
	client dapr.Client

	serviceAddress = getEnvVar("ADDRESS", ":60011")

	sourcePubSubName = getEnvVar("SOURCE_PUBSUB_NAME", "messagebus")
	sourceTopicName  = getEnvVar("SOURCE_TOPIC_NAME", "http")

	targetBindingName = getEnvVar("TARGET_BINDING", "http-binding")
)

func main() {
	// create Dapr service
	s, err := daprd.NewService(serviceAddress)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	c, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("failed to create Dapr client: %v", err)
	}
	client = c
	defer client.Close()

	// add handler to the service
	sub := &common.Subscription{PubsubName: sourcePubSubName, Topic: sourceTopicName}
	s.AddTopicEventHandler(sub, eventHandler)

	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	logger.Printf("Event - PubsubName:%s, Topic:%s, ID:%s", e.PubsubName, e.Topic, e.ID)

	d, ok := e.Data.([]byte)
	if !ok {
		return false, errors.Errorf("invalid event data type: %T", e.Data)
	}

	// content := &dapr.BindingInvocation{
	// 	Data: d,
	// 	Metadata: map[string]string{
	// 		"record-id":       e.ID,
	// 		"conversion-time": time.Now().UTC().Format(time.RFC3339),
	// 	},
	// 	Name:      targetBindingName,
	// 	Operation: "create",
	// }

	// if err := client.InvokeOutputBinding(ctx, content); err != nil {
	// 	return true, errors.Wrap(err, "error invoking target binding")
	// }

	return false, nil
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
