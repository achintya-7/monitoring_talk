FROM caddy:latest

COPY ./infra/alloy-sidecar-caddy/config/Caddyfile /etc/caddy/Caddyfile
