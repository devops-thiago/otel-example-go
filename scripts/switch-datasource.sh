#!/bin/bash

# Script to switch Grafana datasource configuration
# Usage: ./scripts/switch-datasource.sh [mimir|prometheus]

set -e

DEFAULT_DATASOURCE=${1:-mimir}

case $DEFAULT_DATASOURCE in
  mimir)
    echo "Setting Mimir as default datasource..."
    export MIMIR_URL="http://mimir:9009/prometheus"
    export PROMETHEUS_URL="http://prometheus:9090"
    export TEMPO_URL="http://tempo:3200"
    export LOKI_URL="http://loki:3100"
    ;;
  prometheus)
    echo "Setting Prometheus as default datasource..."
    export MIMIR_URL="http://mimir:9009/prometheus"
    export PROMETHEUS_URL="http://prometheus:9090"
    export TEMPO_URL="http://tempo:3200"
    export LOKI_URL="http://loki:3100"
    # Note: You would need to update the datasources.yaml to make Prometheus default
    echo "Warning: You need to manually update datasources.yaml to make Prometheus the default"
    ;;
  *)
    echo "Usage: $0 [mimir|prometheus]"
    echo "  mimir      - Use Mimir as default datasource (recommended)"
    echo "  prometheus - Use Prometheus as default datasource"
    exit 1
    ;;
esac

echo "Datasource configuration:"
echo "  MIMIR_URL: $MIMIR_URL"
echo "  PROMETHEUS_URL: $PROMETHEUS_URL"
echo "  TEMPO_URL: $TEMPO_URL"
echo "  LOKI_URL: $LOKI_URL"

echo ""
echo "To apply changes, restart Grafana:"
echo "  docker-compose restart grafana"
