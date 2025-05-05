FROM gcr.io/distroless/static-debian12:nonroot

COPY prometheus-rds-exporter /app/

HEALTHCHECK NONE

ENTRYPOINT ["/app/prometheus-rds-exporter"]
