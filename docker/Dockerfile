FROM golang:1.15.7 as builder

ARG APP_VERSION=dev

RUN apt-get update && apt-get install -y upx

WORKDIR /src/
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-w -s -X 'main.version=$(APP_VERSION)'" -o service main.go
RUN upx service

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /src/service .

ENTRYPOINT ["./service"]