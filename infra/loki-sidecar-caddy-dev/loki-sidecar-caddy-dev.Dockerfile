FROM caddy:latest

COPY ./infra/loki-sidecar-caddy-dev/config/Caddyfile /etc/caddy/Caddyfile