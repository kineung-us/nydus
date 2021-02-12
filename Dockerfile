FROM golang:1.15.7 as builder

WORKDIR /src/

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o service main.go
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /src/service .

ENTRYPOINT ["./service"]