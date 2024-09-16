FROM grafana/grafana:latest

# Copy configuration files
COPY ./infra/grafana/config /etc/grafana-config
COPY ./infra/grafana/datasources /etc/grafana/provisioning/datasources
COPY ./infra/grafana/dashboards-provisioning /etc/grafana/provisioning/dashboards
COPY ./infra/grafana/dashboards /var/lib/grafana/dashboards
COPY ./infra/grafana/plugins /var/lib/grafana/plugins
COPY ./infra/grafana/plugins-provisioning /etc/grafana/provisioning/plugins

# Set environment variables
ENV GRAFANA_CONFIG_FILE="grafana.ini"
ENV GRAFANA_CONFIG_PATH="/etc/grafana-config"
ENV GRAFANA_DASHBOARDS_PATH="/var/lib/grafana/dashboards"
ENV GRAFANA_DASHBOARDS_PROVISIONING_PATH="/etc/grafana/provisioning/dashboards"
ENV GRAFANA_DATASOURCES_PATH="/etc/grafana/provisioning/datasources"
ENV GRAFANA_HOME_PATH="/usr/share/grafana"
ENV GRAFANA_PLUGINS_PATH="/var/lib/grafana/plugins"
ENV GRAFANA_PLUGINS_PROVISIONING_PATH="/etc/grafana/provisioning/plugins"
ENV GF_SECURITY_ADMIN_USER=achintya
ENV GF_SECURITY_ADMIN_PASSWORD=password

# Fix permissions
USER root
RUN chown -R grafana:root /etc/grafana-config /etc/grafana /var/lib/grafana \
    && chmod -R 755 /etc/grafana-config /etc/grafana /var/lib/grafana

# Switch back to the grafana user
USER grafana

# Use the new command format for starting Grafana
ENTRYPOINT ["grafana", "server", \
            "--homepath=/usr/share/grafana", \
            "--config=/etc/grafana-config/grafana.ini"]