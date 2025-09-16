# Observability Helm Chart

This Helm chart provides a complete observability stack for MediSupply microservices running on Istio service mesh.

## Components

- **Grafana Dashboard**: Pre-configured dashboard for monitoring MediSupply services
- **Prometheus Configuration**: Scraping configuration for service metrics
- **Jaeger Tracing**: Distributed tracing configuration
- **Kiali**: Service mesh observability and management
- **ServiceMonitor**: Optional Prometheus Operator integration

## Prerequisites

- Kubernetes cluster with Istio installed
- Istio namespace (`istio-system`) created
- MediSupply services deployed in the configured namespace

## Installation

```bash
# Install the chart
helm install observability ./k8s/observability -n istio-system

# Install with custom values
helm install observability ./k8s/observability -n istio-system -f custom-values.yaml

# Upgrade
helm upgrade observability ./k8s/observability -n istio-system
```

## Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.namespace` | Istio system namespace | `istio-system` |
| `global.medisupplyNamespace` | MediSupply services namespace | `medisupply` |
| `grafana.enabled` | Enable Grafana dashboard | `true` |
| `prometheus.enabled` | Enable Prometheus configuration | `true` |
| `jaeger.enabled` | Enable Jaeger tracing | `true` |
| `kiali.enabled` | Enable Kiali configuration | `true` |
| `serviceMonitor.enabled` | Use ServiceMonitor (requires Prometheus Operator) | `false` |

### ServiceMonitor vs ConfigMap

By default, the chart uses ConfigMap-based Prometheus configuration. To use ServiceMonitor resources:

1. Install Prometheus Operator
2. Set `serviceMonitor.enabled: true`
3. Set `prometheus.enabled: false`

### Custom Services

To monitor additional services, add them to `prometheus.config.services`:

```yaml
prometheus:
  config:
    services:
      - name: my-service
        metricsPath: /metrics
        port: http
        interval: 30s
```

## Usage

After installation:

1. **Grafana**: Access dashboards through your Grafana instance
2. **Kiali**: Access service mesh visualization at the configured URL
3. **Jaeger**: View distributed traces through Jaeger UI
4. **Prometheus**: Metrics are automatically scraped from configured services

## Troubleshooting

- Ensure Istio is properly installed and configured
- Verify that services expose metrics on the configured endpoints
- Check that namespaces exist before installation
- For ServiceMonitor issues, verify Prometheus Operator is installed

## Examples

See the `k8s/guia/observability/` directory for the original configuration files that inspired this chart.