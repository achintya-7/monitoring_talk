FROM grafana/mimir:latest

COPY ./infra/mimir/config/config.yaml /etc/mimir/config.yaml

ENTRYPOINT ["/bin/mimir", "-config.file=/etc/mimir/config.yaml", "-config.expand-env=true"]
