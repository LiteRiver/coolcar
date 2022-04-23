FROM golang:1.18.1-alpine as builder
WORKDIR /go/src/coolcar/server
COPY . /go/src/coolcar/server
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go install ./gateway/...

FROM alpine:3.15
COPY --from=builder /go/bin/gateway /bin/gateway
EXPOSE 8081
ENTRYPOINT ["/bin/gateway"]