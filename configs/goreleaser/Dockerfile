FROM alpine:3

ARG USER=exporter
ARG HOME=/app

RUN addgroup -g 1001 -S exporter \
    && adduser --home /app -u 1001 -S exporter -G exporter \
    && mkdir -p /app \
    && chown exporter:exporter -R /app

WORKDIR $HOME
USER $USER

EXPOSE 9043

COPY prometheus-rds-exporter /app/
ENTRYPOINT ["/app/prometheus-rds-exporter"]
