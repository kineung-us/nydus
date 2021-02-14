package main

import (
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setHost(t *testing.T) {
	assert := assert.New(t)

	root := "http://localhost:5500/addpath"

	u, _ := url.Parse("http://test:5000/test?how=to")
	k, _ := url.Parse("http://localhost:5500/addpath/test?how=to")

	assert.Equal(k.String(),
		setHost(root, u), "they should be equal")
}

func Test_unmCloudEvnet(t *testing.T) {
	assert := assert.New(t)
	jsn := `{
		"source": "nydus",
		"datacontenttype": "application/json",
		"data": {
			"callback": "http://localhost",
			"meta": null,
			"order": {
				"method": "POST",
				"url": "http://localhost:5000/publish/req-service/3",
				"headers": {
					"Accept": "application/json, text/xml, application/xml, */*",
					"Accept-Encoding": "deflate, gzip",
					"Content-Length": "16",
					"Content-Type": "application/json",
					"Host": "localhost:5000",
					"User-Agent": "libcurl/7.64.1 r-curl/4.3 httr/1.4.2"
				},
				"body": {
					"evnet": "test"
				}
			}
		},
		"time": "2021-02-14T14:24:45Z",
		"traceid": "00-f95d43e0afecca9a7e6e0f072001bbf4-b24c67605e77eb80-01",
		"topic": "req-service",
		"pubsubname": "pubsub",
		"id": "9b3d02e1-e425-4cb5-91f1-593aab87ae00"
	}`

	ce := customEvent{}
	log.Println(jsn)
	log.Println(ce)

	if err := json.Unmarshal([]byte(jsn), &ce); err != nil {
		fmt.Println("error")
		log.Fatalln(err)
	}
	log.Println(ce)
	log.Println(ce.Data)
	log.Println(ce.Data.Order)
	assert.Equal(jsn,
		ce, "they should be equal")
}
