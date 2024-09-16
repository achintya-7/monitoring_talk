FROM caddy:latest

COPY ./infra/alloy-sidecar-caddy-dev/config/Caddyfile /etc/caddy/Caddyfile
