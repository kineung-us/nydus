package body

import (
	"encoding/json"
	"fmt"
	"net/url"
	"nydus/pkg/env"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func Unmarshal(raw []byte, ct string) (map[string]interface{}, error) {
	b := map[string]interface{}{}
	switch {
	case strings.Contains(ct, "json"):
		if err := json.Unmarshal(raw, &b); err != nil {
			return nil, err
		}
	case strings.Contains(ct, "xml"):
		log.Info().Str("xmlraw", string(raw)).Send()
		j, err := mxj.NewMapXml(raw)
		if err != nil {
			return nil, err
		}
		b = j
	case strings.Contains(ct, "x-www-form-urlencoded"):
		log.Info().Str("body", string(raw)).Send()
		ss := strings.Split(string(raw), "&")
		for _, s := range ss {
			kv := strings.Split(s, "=")
			if len(kv) == 1 {
				b[kv[0]] = nil
			} else {
				b[kv[0]] = kv[1]
			}
		}
	default:
		b["string"] = string(raw)
	}
	return b, nil
}

func Marshal(d map[string]interface{}, ct string) ([]byte, error) {
	b := []byte{}
	switch {
	case strings.Contains(ct, "json"):
		j, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		b = j
	case strings.Contains(ct, "xml"):
		mv := mxj.Map(d)
		xmlValue, err := mv.Xml()
		if err != nil {
			return nil, err
		}
		b = xmlValue
	case strings.Contains(ct, "x-www-form-urlencoded"):
		r := []string{}
		for k, v := range d {
			if v == nil {
				r = append(r, k)
			} else {
				r = append(r, strings.Join([]string{k, fmt.Sprintf("%v", v)}, "="))
			}
		}
		b = []byte(strings.Join(r, "&"))
	default:
		b = []byte(fmt.Sprintf("%v", d["string"]))
	}
	return b, nil
}

type CustomEvent struct {
	ID              uuid.UUID    `json:"id"`
	Source          string       `json:"source"`
	Type            string       `json:"type"`
	SpecVersion     string       `json:"specversion"`
	DataContentType string       `json:"datacontenttype"`
	Topic           string       `json:"topic"`
	TraceID         string       `json:"traceid,omitempty"`
	Data            *PublishData `json:"data"`
	Time            string       `json:"time"`
}

func (c *CustomEvent) PropTrace() {
	c.Data.Order.Headers["traceparent"] = c.TraceID
}

type PublishData struct {
	Order    *RequestedData         `json:"order"`
	Callback string                 `json:"callback"`
	Meta     map[string]interface{} `json:"meta"`
}

type RequestedData struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

type Message struct {
	ID      string
	Status  string
	Headers map[string]string
	Body    map[string]interface{}
}

type Callback struct {
	Callback string        `json:"callback"`
	ID       uuid.UUID     `json:"id"`
	Response *ResponseData `json:"response"`
}

type ResponseData struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

func NewCustomEvent(pub *PublishData, tid string, targetTopic string) *CustomEvent {
	tz, _ := time.LoadLocation("UTC")
	return &CustomEvent{
		ID:              uuid.New(),
		Source:          "nydus",
		Type:            "com.dapr.event.sent",
		SpecVersion:     "1.0",
		DataContentType: "application/json",
		Data:            pub,
		Topic:           targetTopic,
		TraceID:         tid,
		Time:            time.Now().In(tz).Format(time.RFC3339),
	}
}

func (p *PublishData) UpdateHost(r string) {
	t, _ := url.Parse(p.Order.URL)
	p.Order.URL = setHost(r, t)
}

func setHost(r string, u *url.URL) string {
	t, _ := url.Parse(r)
	u.Scheme = t.Scheme
	u.Host = t.Host + t.Path
	u.Path = strings.ReplaceAll(u.Path, "/publish/"+env.SubscribeTopic, "")
	h, _ := url.PathUnescape(u.String())
	return h
}
