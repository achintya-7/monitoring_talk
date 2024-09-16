FROM grafana/alloy:boringcrypto

COPY ./infra/alloy/config/config.alloy /etc/alloy/config.alloy

EXPOSE 12345
EXPOSE 8027
EXPOSE 4317
EXPOSE 4318

CMD ["run", "/etc/alloy/config.alloy", "--storage.path=/var/lib/alloy/data", "--server.http.listen-addr=0.0.0.0:12345", "--config.extra-args=\"-config.expand-env\""]
