FROM caddy:latest

COPY ./infra/tempo-sidecar-caddy-dev/config/Caddyfile /etc/caddy/Caddyfile
