FROM grafana/tempo:latest

COPY ./infra/tempo/config/tempo.yaml /etc/tempo-config/tempo.yaml

CMD ["-config.file=/etc/tempo-config/tempo.yaml", "-config.expand-env=true"]
