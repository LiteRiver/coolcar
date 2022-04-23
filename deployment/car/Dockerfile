FROM golang:1.18.1-alpine as builder
WORKDIR /go/src/coolcar/server
COPY . /go/src/coolcar/server
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go install ./car/...
RUN go install github.com/grpc-ecosystem/grpc-health-probe

FROM alpine:3.15
COPY --from=builder /go/bin/car /bin/car
COPY --from=builder /go/bin/grpc-health-probe /bin/grpc-health-probe
EXPOSE 8081
ENTRYPOINT ["/bin/car"]