FROM golang:1.23 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

HEALTHCHECK NONE

RUN make build


FROM alpine:3.20

RUN apk add aws-cli curl

ARG USER=exporter
ARG HOME=/app

RUN addgroup -g 1001 -S exporter \
    && adduser --home /app -u 1001 -S exporter -G exporter \
    && mkdir -p /app \
    && chown exporter:exporter -R /app

WORKDIR $HOME
USER $USER

COPY --from=builder /build/prometheus-rds-exporter $HOME/prometheus-rds-exporter

EXPOSE 9043

HEALTHCHECK CMD curl --fail http://localhost:3000 || exit 1

ENTRYPOINT [ "/app/prometheus-rds-exporter" ]
