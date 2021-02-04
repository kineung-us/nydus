# nydus-enter

https://jacking75.github.io/go_channel_howto/

sync로 요청하는 http 연결을 뒤에 pubsub에 보내고 응답을 받을 때 까지 유지하는 역할.
request와 callback 2개의 엔드포인트가 필요함.
request-id 이름으로 callback 엔드포인트에 체널을 보내두고 id=request-id 인 것이 있을 때 해당 체널로 body를 전달하여 result를 응답함.

need to get targetName.
need to add hostIP, targetName.

## endpoint
### /invoke/:target

외부에서 target 으로 http 요청을 하는 것을 모사함.
요청을 받으면 targetName과 자신의 hostIP를 추가하여 pub을 진행.

https://github.com/guiguan/caster#broadcast-a-go-channel

### /callback/:id


## body example

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
    "queryStringParameters":{"city":"Seattle",
    "name":"John"
  },
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