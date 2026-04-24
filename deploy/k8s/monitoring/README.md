# Monitoring — Prometheus + Grafana

Minimal observability layer for the `relax-hub` cluster.

## What runs here

| Component | Replicas | CPU req/lim | RAM req/lim | Storage |
|---|---|---|---|---|
| prometheus-operator | 1 | 50m / 200m | 64Mi / 256Mi | — |
| prometheus | 1 | 200m / 500m | 512Mi / 1Gi | 10 Gi Ceph SSD (7d retention) |
| grafana | 1 | 50m / 200m | 96Mi / 256Mi | 2 Gi Ceph SSD |

Everything sits in namespace `monitoring` (pre-created by VK Cloud).
Alertmanager / node-exporter / kube-state-metrics / control-plane scrapes are all **disabled** — too heavy for a 4 vCPU / 14 GiB single-node cluster. We can flip them on later.

## One-time install

```bash
export KUBECONFIG=.deploy-local/kubeconfig.yaml

# 1. Generate a Grafana admin password and stash it locally (gitignored).
mkdir -p .deploy-local
GRAFANA_PASS=$(openssl rand -hex 24)
echo "$GRAFANA_PASS" > .deploy-local/grafana-admin.txt
chmod 600 .deploy-local/grafana-admin.txt

# 2. Add Helm repo.
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# 3. Install the stack.
helm upgrade --install kps prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  -f deploy/k8s/monitoring/values.yaml \
  --set grafana.adminPassword="$GRAFANA_PASS" \
  --version '~65' \
  --wait --timeout=5m

# 4. Apply the ServiceMonitor for our API.
kubectl apply -f deploy/k8s/monitoring/servicemonitor.yaml
```

## Access (Grafana is NOT publicly exposed)

```bash
kubectl -n monitoring port-forward svc/kps-grafana 3000:80
# open http://localhost:3000
# user:  admin
# pass:  cat .deploy-local/grafana-admin.txt
```

Prometheus UI (rarely needed — Grafana is enough):

```bash
kubectl -n monitoring port-forward svc/kps-kube-prometheus-stack-prometheus 9090:9090
# open http://localhost:9090
```

## Verify scrape target

In Grafana → *Explore* → Prometheus datasource → query:

```
up{job="relax-hub-api"}
```

Should return `1`. If `0` or missing:

```bash
kubectl -n relax-hub get svc relax-hub-api -o jsonpath='{.spec.ports}' | jq
# expect port name "metrics" → 9090

kubectl -n monitoring get servicemonitor relax-hub-api -o yaml
# expect endpoints[0].port == metrics

kubectl -n monitoring logs -l app.kubernetes.io/name=prometheus --tail=50 | grep -i relax-hub
```

## Ready-made dashboards to import

| ID | Name | Notes |
|----|------|-------|
| [10826](https://grafana.com/grafana/dashboards/10826) | Go / chi HTTP | Matches our `http_requests_total` / `http_request_duration_seconds` labels |
| [6671](https://grafana.com/grafana/dashboards/6671) | Go processes | `go_*` runtime metrics (GC, goroutines, memory) |

Import via Grafana UI → *+ → Import → paste ID*. Select the `Prometheus` datasource.

## Upgrade / teardown

```bash
# Upgrade (e.g. after tweaking values.yaml)
helm upgrade kps prometheus-community/kube-prometheus-stack \
  -n monitoring -f deploy/k8s/monitoring/values.yaml --reuse-values

# Teardown (keeps PVCs)
helm uninstall kps -n monitoring
kubectl delete -f deploy/k8s/monitoring/servicemonitor.yaml

# Full wipe
kubectl -n monitoring delete pvc -l app.kubernetes.io/instance=kps
```

## Known gaps / next steps

- **No alerting**: alertmanager disabled. When we wire a channel (email / telegram / VK teams) — re-enable in `values.yaml` with a `config` stanza.
- **No node-level metrics**: nodeExporter off. Add back if we start running out of host resources unexpectedly.
- **DB/Redis/MinIO not scraped**: add prometheus-postgres-exporter, redis_exporter, minio built-in metrics as follow-ups once the base loop is stable.
