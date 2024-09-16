FROM caddy:latest

COPY ./infra/loki-sidecar-caddy/config/Caddyfile /etc/caddy/Caddyfile