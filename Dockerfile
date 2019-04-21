FROM golang:1.12.4-stretch

WORKDIR /go/apm-gateway/
COPY ./go.mod ./go.sum /go/apm-gateway/
RUN go mod download
COPY ./main.go /go/apm-gateway/main.go
COPY ./pkg /go/apm-gateway/pkg/
RUN go test ./... && CGO_ENABLED=0 GOOS=linux go build

FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=0 /go/apm-gateway/apm-gateway /usr/bin/apm-gateway
ENTRYPOINT ["/usr/bin/apm-gateway"]
