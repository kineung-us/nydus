# http_client

메타를 기준으로 요청을 수행.

json to request 같은 동작이 되어야 함.

```json
{
  "http-method": "get",
  "header": {
    "Accept": "application/json, text/xml, application/xml, */*", 
    "Accept-Encoding": "deflate, gzip", 
    "Content-Length": "15", 
    "Content-Type": "application/json", 
    "Host": "httpbin.org", 
    "User-Agent": "libcurl/7.64.1 r-curl/4.3 httr/1.4.2", 
    "X-Amzn-Trace-Id": "Root=1-601d053f-12675bab033b812a1117ac76"
  }, 
  "source": {
    "hostIP": "",
    "origin": ""
    "": ""
  },
  "target": {
    "root": "https://httpbin.org",
    "path": "/get"
    "params": {
      "test": "data" 
    }
  },
  "data": {}
}
```

## example

### go http package request struct

```go
type Request struct {
    Method string
    URL *url.URL
    Proto      string // e.g. "HTTP/1.0"
    Header Header
    Body io.ReadCloser
    ContentLength int64
    Host string
    Form url.Values
    PostForm url.Values // Go 1.1
    MultipartForm *multipart.Form
    Trailer Header
    RemoteAddr string
    RequestURI string
}


// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
type URL struct {
    Scheme      string
    Opaque      string    // encoded opaque data
    User        *Userinfo // username and password information
    Host        string    // host or host:port
    Path        string    // path (relative paths may omit leading slash)
    RawPath     string    // encoded path hint (see EscapedPath method); added in Go 1.5
    ForceQuery  bool      // append a query ('?') even if RawQuery is empty; added in Go 1.7
    RawQuery    string    // encoded query values, without '?'
    Fragment    string    // fragment for references, without '#'
    RawFragment string    // encoded fragment hint (see EscapedFragment method); added in Go 1.15
}

type Header map[string][]string

type Userinfo struct {
	username    string
	password    string
	passwordSet bool
}

type Response struct {
    Status     string // e.g. "200 OK"
    StatusCode int    // e.g. 200
    Proto      string // e.g. "HTTP/1.0"
    Header Header
    Body io.ReadCloser
    ContentLength int64
    Uncompressed bool // Go 1.7
    Trailer Header
    Request *Request
    TLS *tls.ConnectionState // Go 1.3
}
```


### httpbin

```json
{
  "args": {
    "test": "test"
  }, 
  "data": "{\"body\":\"test\"}", 
  "files": {}, 
  "form": {}, 
  "headers": {
    "Accept": "application/json, text/xml, application/xml, */*", 
    "Accept-Encoding": "deflate, gzip", 
    "Content-Length": "15", 
    "Content-Type": "application/json", 
    "Host": "httpbin.org", 
    "User-Agent": "libcurl/7.64.1 r-curl/4.3 httr/1.4.2", 
    "X-Amzn-Trace-Id": "Root=1-601d053f-12675bab033b812a1117ac76"
  }, 
  "json": {
    "body": "test"
  }, 
  "origin": "223.38.54.28", 
  "url": "https://httpbin.org/post?test=test"
}
```

### cloud event spec

https://github.com/cloudevents/spec/blob/v1.0.1/spec.md

```json
{
    "specversion" : "1.0",
    "type" : "com.github.pull_request.opened",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\"xml\"/>"
}
```

### aws lambda input

https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-create-api-as-simple-proxy-for-lambda.html

```json
{
  "message":"Good evening, John of Seattle. Happy Thursday!", 
  "input":{
    "resource":"/helloworld",
    "path":"/helloworld",
    "httpMethod":"POST",
    "headers":{
      "Accept":"*/*",
      "content-type":"application/json",
      "day":"Thursday",
      "Host":"r275xc9bmd.execute-api.us-east-1.amazonaws.com",
      "User-Agent":"curl/7.64.0",
      "X-Amzn-Trace-Id":"Root=1-1a2b3c4d-a1b2c3d4e5f6a1b2c3d4e5f6",
      "X-Forwarded-For":"72.21.198.64",
      "X-Forwarded-Port":"443",
      "X-Forwarded-Proto":"https"
    },
    "multiValueHeaders":{"Accept":["*/*"],
    "content-type":["application/json"],
    "day":["Thursday"],
    "Host":["r275xc9bmd.execute-api.us-east-1.amazonaws.com"],
    "User-Agent":["curl/0.0.0"],
    "X-Amzn-Trace-Id":["Root=1-1a2b3c4d-a1b2c3d4e5f6a1b2c3d4e5f6"],
    "X-Forwarded-For":["11.22.333.44"],
    "X-Forwarded-Port":["443"],
    "X-Forwarded-Proto":["https"]},
    "queryStringParameters":{"city":"Seattle","name":"John"},
  "multiValueQueryStringParameters":{
    "city":["Seattle"],
    "name":["John"]
  },
  "pathParameters":null,
  "stageVariables":null,
  "requestContext":{
    "resourceId":"3htbry",
    "resourcePath":"/helloworld",
    "htt* Connection #0 to host r275xc9bmd.execute-api.us-east-1.amazonaws.com left intact pMethod":"POST",
    "extendedRequestId":"a1b2c3d4e5f6g7h=",
    "requestTime":"20/Mar/2019:20:38:30 +0000",
    "path":"/test/helloworld",
    "accountId":"123456789012",
    "protocol":"HTTP/1.1",
    "stage":"test",
    "domainPrefix":"r275xc9bmd",
    "requestTimeEpoch":1553114310423,
    "requestId":"test-invoke-request",
    "identity":{"cognitoIdentityPoolId":null,
      "accountId":null,
      "cognitoIdentityId":null,
      "caller":null,
      "sourceIp":"test-invoke-source-ip",
      "accessKey":null,
      "cognitoAuthenticationType":null,
      "cognitoAuthenticationProvider":null,
      "userArn":null,
      "userAgent":"curl/0.0.0","user":null
    },
    "domainName":"r275xc9bmd.execute-api.us-east-1.amazonaws.com",
    "apiId":"r275xc9bmd"
  },
  "body":"{ \"time\": \"evening\" }",
  "isBase64Encoded":false
  }
}
```