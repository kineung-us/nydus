package call_test

import (
	"fmt"
	"nydus/pkg/body"
	"nydus/pkg/call"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestUnmarshal(t *testing.T) {
	in := "https://den.prd.sktchatbot.co.kr/prompts?promptID=PromptFAQ_010%EC%A0%84%ED%99%98%EB%B2%88%ED%98%B8%EB%B3%80%EA%B2%BD"
	fmt.Println(in)
	tem := body.RequestedData{
		Method:  "get",
		URL:     in,
		Headers: map[string]string{},
		Body:    nil,
	}

	res, _ := call.RequesttoTarget(&tem)
	actual, _ := body.Unmarshal(res.Body.([]byte), res.Headers["Content-Type"])
	expected := []interface{}([]interface{}{map[string]interface{}{"Test": "Test"}, map[string]interface{}{"Test2": "Test2"}})
	assert.Equal(t, expected, actual, "기대값과 결과값이 다릅니다.")
}

func TestGet(t *testing.T) {
	addr := "https://hydra.dev.sktchatbot.co.kr"
	st, _, _ := fasthttp.GetTimeout(nil, addr, 3*time.Second)
	assert.Equal(t, 200, st, "기대값과 결과값이 다릅니다.")
}
