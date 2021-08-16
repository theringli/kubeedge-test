FROM golang:1.13-alpine3.11 AS builder

COPY . /go/src/github.com/subpathdev/CpuTempMqttClient
RUN cd /go/src/github.com/subpathdev/CpuTempMqttClient; go build -o /usr/local/bin/CpuTempMqttClient

FROM alpine:3.11

RUN apk add --no-cache lm_sensors
COPY --from=builder /usr/local/bin/CpuTempMqttClient /usr/local/bin/CpuTempMqttClient

ENTRYPOINT ["/usr/local/bin/CpuTempMqttClient"]

