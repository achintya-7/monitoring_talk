FROM caddy:latest

COPY ./infra/tempo-sidecar-caddy/config/Caddyfile /etc/caddy/Caddyfile
