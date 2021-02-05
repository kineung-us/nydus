package main

import (
	"context"
	"log"
	"os"
	"strings"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/guiguan/caster"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "", 0)
	client dapr.Client

	serviceAddress = getEnvVar("ADDRESS", ":5000")

	pubSubName   = getEnvVar("SOURCE_PUBSUB_NAME", "messagebus")
	requestTopic = getEnvVar("SOURCE_TOPIC_NAME", "req-service")
	resultTopic  = getEnvVar("RESULT_TOPIC_NAME", "res-service")

	targetRoot = getEnvVar("TARGET_ROOT", "app-id")
)

func main() {
	// create Dapr service
	s := daprd.NewService(serviceAddress)

	p := caster.New(nil)
	c, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("failed to create Dapr client: %v", err)
	}
	client = c
	defer client.Close()

	s.AddServiceInvocationHandler("invoke", invoke(c, p))
	s.AddServiceInvocationHandler("callback", callback(p))

	reqsub := &common.Subscription{
		PubsubName: pubSubName,
		Topic:      requestTopic,
	}
	s.AddTopicEventHandler(reqsub, reqSub(c))

	ressub := &common.Subscription{
		PubsubName: pubSubName,
		Topic:      resultTopic,
	}
	s.AddTopicEventHandler(ressub, resultSub(p))
	// start the server to handle incoming events
	log.Printf("starting server at %s...", serviceAddress)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func invoke(c dapr.Client, p *caster.Caster) func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		ch, ok := p.Sub(nil, 1)
		if !ok {
			return nil, errors.Errorf("invalid event data type")
		}
		defer p.Unsub(ch)

		logger.Printf(
			"Invocation (ContentType:%s, Verb:%s, QueryString:%s, Data:%s)",
			in.ContentType, in.Verb, in.QueryString, string(in.Data),
		)

		if err := c.PublishEvent(ctx, pubSubName, pubSubName, b); err != nil {
			return nil, errors.Wrap(err, "error publishing content")
		}

		// TODO 양식 맞춰서 고쳐야 함
		for m := range ch {
			t := m.(message).ID
			if c.Params("id") == t {
				body = m.(message).Body
				break
			}
		}
		return c.Send(body)
	}
}

func callback(p *caster.Caster) func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

		logger.Printf(
			"Invocation (ContentType:%s, Verb:%s, QueryString:%s, Data:%s)",
			in.ContentType, in.Verb, in.QueryString, string(in.Data),
		)

		p.Pub(message{
			ID:   c.Params("id"),
			Body: c.Body(),
		})
		return c.SendString("send!")
	}
}

func resultSub(p *caster.Caster) func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	return func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		// ip가 호스트와 같은지 확인
		// 같으면 caster로 보냄
		// 다르면 post로 ip에 요청함
		// 이러고 종료
	}
}

func reqSub(c dapr.Client) func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	return func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		// req를 post로 url에 전달
		// targetRoot + path 등등 수행
	}
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
