package http_client

type HTTPOrder struct {
	Method  string      `json:method`
	URL     string      `json:url`
	Headers interface{} `json:headers`
	Body    interface{} `json:body`
}
