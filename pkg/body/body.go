package body

import (
	"errors"
	"fmt"
	"net/url"
	"nydus/pkg/env"
	"strings"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var (
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
	subTopic  = env.SubscribeTopic
	dftstring = env.DFTtoString
	xmlstring = env.XMLtoString
	tz, _     = time.LoadLocation("UTC")
)

func Unmarshal(raw []byte, ct string) (interface{}, error) {
	var b interface{}
	switch {
	case string(raw) == "":
		b = raw
	case raw == nil:
		b = nil
	case strings.Contains(ct, "json"):
		log.Debug().Str("json", string(raw)).Send()
		d := map[string]interface{}{}
		if err := json.Unmarshal(raw, &d); err != nil {
			dd := []interface{}{}
			if err := json.Unmarshal(raw, &dd); err != nil {
				return nil, err
			}
			b = dd
		} else {
			b = d
		}
	case strings.Contains(ct, "xml") || strings.Contains(ct, "html"):
		log.Debug().Bool("xmlstring", xmlstring).Send()
		log.Debug().Str("xmlraw", string(raw)).Send()
		if xmlstring {
			b = string(raw)
		} else {
			j, err := mxj.NewMapXml(raw)
			if err != nil {
				return nil, err
			}
			b = map[string]interface{}(j)
		}
	case strings.Contains(ct, "x-www-form-urlencoded"):
		log.Debug().Str("form", string(raw)).Send()
		ss := strings.Split(string(raw), "&")
		f := map[string]interface{}{}
		for _, s := range ss {
			kv := strings.Split(s, "=")
			if len(kv) == 1 {
				f[kv[0]] = nil
			} else {
				f[kv[0]] = kv[1]
			}
		}
		b = f
	default:
		log.Debug().Bool("dftstring", dftstring).Send()
		log.Debug().Str("string", string(raw)).Send()
		if xmlstring {
			b = string(raw)
		} else {
			b = ""
		}
	}
	return b, nil
}

func Marshal(d interface{}, ct string) ([]byte, error) {
	b := []byte{}
	switch {
	case strings.Contains(ct, "json"):
		j, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		b = j
	case strings.Contains(ct, "xml"):
		switch dt := d.(type) {
		case map[string]interface{}:
			dm := dt
			mv := mxj.Map(dm)
			xmlValue, err := mv.Xml()
			if err != nil {
				return nil, err
			}
			b = xmlValue
		case string:
			b = []byte(dt)
		}

	case strings.Contains(ct, "x-www-form-urlencoded"):
		r := []string{}
		for k, v := range d.(map[string]interface{}) {
			if v == nil {
				r = append(r, k)
			} else {
				r = append(r, strings.Join([]string{k, fmt.Sprintf("%v", v)}, "="))
			}
		}
		b = []byte(strings.Join(r, "&"))
	case d == nil:
		b = nil
	default:
		b = []byte(fmt.Sprintf("%v", d))
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
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

type Message struct {
	ID      string
	Status  string
	Headers map[string]string
	Body    interface{}
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

func (p *PublishData) UpdateHost(r string) error {
	if p.Order == nil {
		return errors.New("PublishData.Order cannot be nil")
	}
	t, err := url.Parse(p.Order.URL)
	if err != nil {
		return err
	}
	p.Order.URL = SetHost(r, t)
	log.Debug().Str("Location", "Before URLencod").Interface("p.Order.URL", p.Order.URL).Send()
	p.Order.URL = URLencode(p.Order.URL)
	log.Debug().Str("Location", "After URLencod").Interface("p.Order.URL", p.Order.URL).Send()
	return nil
}

func SetHost(r string, u *url.URL) string {
	t, _ := url.Parse(r)
	u.Scheme = t.Scheme
	u.Host = t.Host + t.Path
	u.Path = strings.ReplaceAll(u.Path, "/publish/"+subTopic, "")
	h, _ := url.PathUnescape(u.String())
	return h
}

func URLencode(h string) string {
	t, _ := url.Parse(h)
	if t.Query().Encode() == "" {
		return t.Scheme + "://" + t.Host + t.Path
	}
	return t.Scheme + "://" + t.Host + t.Path + "?" + t.Query().Encode()
}
