FROM golang:1.15.7 as builder

RUN apt-get update && apt-get install -y upx

WORKDIR /src/
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o service main.go
RUN upx service

FROM gcr.io/distroless/static:nonroot
# FROM bitnami/minideb:buster
COPY --from=builder /src/service .

ENTRYPOINT ["./service"]