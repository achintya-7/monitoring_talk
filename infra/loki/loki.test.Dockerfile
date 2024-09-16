FROM grafana/loki:latest

COPY ./infra/loki/config/loki-config.yaml /etc/loki/local-config.yaml

CMD ["-config.file=/etc/loki/local-config.yaml", "-config.expand-env=true"]
