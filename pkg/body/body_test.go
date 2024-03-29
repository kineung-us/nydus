package body_test

import (
	"fmt"
	"net/url"
	"nydus/pkg/body"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func TestUnmarshal(t *testing.T) {
	tem := `{"Test": "Test"}`

	actual, _ := body.Unmarshal([]byte(tem), "application/json")
	expected1 := map[string]interface{}{"Test": "Test"}
	assert.Equal(t, expected1, actual, "기대값과 결과값이 다릅니다.")

	tem = `[{"Test": "Test"}, {"Test2": "Test2"}]`

	actual, _ = body.Unmarshal([]byte(tem), "application/json")
	fmt.Println(actual)
	mar, _ := body.Marshal(actual, "application/json")
	fmt.Println(string(mar))
	expected2 := []interface{}([]interface{}{map[string]interface{}{"Test": "Test"}, map[string]interface{}{"Test2": "Test2"}})
	assert.Equal(t, expected2, actual, "기대값과 결과값이 다릅니다.")

	actual, _ = body.Unmarshal([]byte(""), "")
	fmt.Println(actual)
	expected3 := ""
	assert.Equal(t, expected3, actual, "기대값과 결과값이 다릅니다.")
}

func TestUpdateHost(t *testing.T) {
	tem := `{"id":"531fe07d-05df-48d8-b868-e2a6d3450020","source":"botjosa","type":"com.dapr.event.sent","specversion":"1.0","datacontenttype":"application/json","topic":"botjosa","data":{"order":null,"callback":"","meta":null},"time":""}`
	ce := body.CustomEvent{}

	json.Unmarshal([]byte(tem), &ce)
	fmt.Println(ce.Data.Order)

	if err := ce.Data.UpdateHost("localhost:8080"); err != nil {
		assert.EqualError(t, err, "PublishData.Order cannot be nil")
	}
}

func TestUrlencode(t *testing.T) {
	tem := `{"id":"531fe07d-05df-48d8-b868-e2a6d3450020","source":"botjosa","type":"com.dapr.event.sent","specversion":"1.0","datacontenttype":"application/json","topic":"botjosa","data":{"order": {"url": "https://www.hello.com/path1/path2?q=PromptFAQ_010전환번호변경", "method": "get"},"callback":"","meta":null},"time":""}`
	ce := body.CustomEvent{}
	json.Unmarshal([]byte(tem), &ce)
	ce.Data.UpdateHost("http://localhost:8080")
	fmt.Println(ce.Data.Order)
	assert.Equal(t, "tem", ce.Data.Order, "기대값과 결과값이 다릅니다.")
}

func TestUrlencodeAddress(t *testing.T) {
	tem := `{"id":"531fe07d-05df-48d8-b868-e2a6d3450020","source":"botjosa","type":"com.dapr.event.sent","specversion":"1.0","datacontenttype":"application/json","topic":"botjosa","data":{"order": {"url": "https://www.hello.com/publish/echo/path1/path2?q=PromptFAQ_010전환번호변경", "method": "get"},"callback":"","meta":null},"time":""}`
	ce := body.CustomEvent{}
	json.Unmarshal([]byte(tem), &ce)
	ce.Data.UpdateHost("http://localhost:8080")
	fmt.Println(ce.Data.Order)
	assert.Equal(t, "tem", ce.Data.Order, "기대값과 결과값이 다릅니다.")
}

func TestSethost(t *testing.T) {
	tem, _ := url.Parse("https://localhost:5000/publish/")
	tt := body.SetHost("http://localhost:8080", tem)
	assert.Equal(t, "http://localhost:8080/publish/", tt, "기대값과 결과값이 다릅니다.")
}

func Test_URLencode(t *testing.T) {
	tem := "http://localhost:8000/predict?text=%EC%9D%B4%EB%B2%88+%EB%8B%AC+%EC%9A%94%EA%B8%88+%EC%96%BC%EB%A7%88+%EB%82%98%EC%99%94%EC%96%B4%3F"
	te := body.URLencode(tem)
	assert.Equal(t, "tem", te, "기대값과 결과값이 다릅니다.")
}
