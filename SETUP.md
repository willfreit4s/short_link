# Setup Guide

This document explains how to run the Short Link service locally in a Kubernetes cluster using Kind, along with PostgreSQL, Redis, monitoring and troubleshooting steps.

## Prerequisites

Make sure the following tools are installed and available:

- Docker
- kubectl
- kind
- make (optional, only for convenience aliases)
- Helm (for monitoring stack installation)

Run all commands from the project root.

## 1. Create the cluster

The file `k8s/kind-config.yaml` creates a cluster named `short-link` with one control-plane node and two workers:

```bash
kind create cluster --config k8s/kind-config.yaml
kubectl cluster-info --context kind-short-link
kubectl get nodes
```

If the cluster already exists, skip the first command. To recreate it from scratch:

```bash
kind delete cluster --name short-link
kind create cluster --config k8s/kind-config.yaml
```

## 2. Build and load the image

The Dockerfile uses a multi-stage build and produces a small image with a non-root user:

```bash
docker build -t short-link:local .
kind load docker-image short-link:local --name short-link
```

The deployment uses `imagePullPolicy: IfNotPresent`, so the cluster will use the locally loaded image.

## 3. Create the namespace

```bash
kubectl apply -f k8s/namespace.yaml
kubectl get namespace short-link
```

## 4. Apply API configuration

Apply the ConfigMap and Secret separately:

```bash
kubectl apply -f k8s/api/configmap.yaml
kubectl apply -f k8s/api/secret.yaml
```

> `k8s/api/secret.yaml` contains development credentials. Do not use these values in production; replace them with a proper secret management mechanism.

## 5. Start PostgreSQL

This project offers two ways to run PostgreSQL:

- `k8s/postgres/statefulset.yaml`: recommended production-like option with a 5 GiB PVC.
- `k8s/postgres/deployment.yaml`: simple disposable option using `emptyDir`.

Use only one of them. Do not apply both together because they both create a resource named `postgres`.

Recommended environment:

```bash
kubectl apply -f k8s/postgres/service.yaml
kubectl apply -f k8s/postgres/statefulset.yaml
kubectl rollout status statefulset/postgres -n short-link --timeout=180s
kubectl get svc postgres -n short-link
```

For disposable environment:

```bash
kubectl apply -f k8s/postgres/deployment.yaml
kubectl rollout status deployment/postgres -n short-link --timeout=180s
```

The Service must exist before the API so that the hostname `postgres` resolves correctly inside the cluster.

## 6. Start Redis

Apply the Service first and then the Deployment:

```bash
kubectl apply -f k8s/redis/service.yaml
kubectl apply -f k8s/redis/deployment.yaml
kubectl rollout status deployment/redis -n short-link --timeout=180s
kubectl get svc redis -n short-link
```

Redis uses `emptyDir`, so data is lost when the pod is removed.

## 7. Run the manual migration

Migrations are executed by the user from the terminal, outside of a Kubernetes Job. First confirm PostgreSQL is ready:

```bash
kubectl get pod -n short-link -l app=postgres
kubectl port-forward -n short-link service/postgres 5432:5432
```

Keep the port-forward open and, in another terminal, run the migration using the local `migrate` binary:

```bash
migrate -path sql/migrations \
  -database 'postgres://postgres:postgres@localhost:5432/short_link?sslmode=disable' up
```

The command should report that migration `1/u init` has been applied. If you prefer not to install the binary locally, you can run the same command via Docker:

```bash
docker run --rm --network host \
  -v "$PWD/sql/migrations:/migrations:ro" \
  migrate/migrate:v4.18.3 \
  -path=/migrations \
  -database 'postgres://postgres:postgres@localhost:5432/short_link?sslmode=disable' up
```

The `--network host` flag allows the Docker container to access the port exposed by `kubectl port-forward`. The migration must finish before starting the API.

## 8. Start the API

```bash
kubectl apply -f k8s/api/service.yaml
kubectl apply -f k8s/api/deployment.yaml
kubectl rollout status deployment/short-link -n short-link --timeout=180s
kubectl get pods,svc -n short-link -o wide
```

The `service.yaml` creates the internal `short-link` Service, and `deployment.yaml` creates 3 API replicas using the image loaded in step 2.

To access the API locally, forward the service to port 8080:

```bash
kubectl port-forward -n short-link service/short-link 8080:8080
```

In another terminal:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl -i -X POST http://localhost:8080/api/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://kubernetes.io"}'
```

The creation response contains `short_url`. The redirect endpoint is `/r/<hash>`.

## 9. HPA, Gateway API and metrics

### HPA

HPA requires the Metrics Server. On a local Kind cluster, install it before applying the manifest if it is not already available:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl patch deployment metrics-server -n kube-system --type=json \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
kubectl rollout status deployment/metrics-server -n kube-system --timeout=180s
kubectl apply -f k8s/api/hpa.yaml
kubectl get hpa -n short-link
```

The `TARGETS` value may remain as `unknown` for a while while Metrics Server collects data.

### Gateway API

`k8s/api/gatewayclass.yaml`, `gateway.yaml` and `httproute.yaml` require a Gateway API controller. This project uses Envoy Gateway; install a compatible version as described in the official documentation before applying the manifests:

```bash
kubectl apply -f k8s/api/gatewayclass.yaml
kubectl apply -f k8s/api/gateway.yaml
kubectl apply -f k8s/api/httproute.yaml
kubectl get gatewayclass,gateway,httproute -n short-link
```

Without Envoy Gateway, these objects may be accepted by the Kubernetes API server but will not create a reachable endpoint.

### Prometheus Operator

The `ServiceMonitor` depends on the CRD provided by the Prometheus Operator. For this project, the recommended approach is to install the `kube-prometheus-stack` with Helm. The release name must be `monitoring`, because `k8s/monitoring/short-link-servicemonitor.yaml` uses the label `release: monitoring`.

Verify Helm is installed and add the chart repository:

```bash
helm version
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

Install the stack in the `monitoring` namespace:

```bash
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

Wait for the main components:

```bash
kubectl wait --for=condition=Available \
  deployment/monitoring-kube-prometheus-operator \
  -n monitoring --timeout=300s

kubectl get pods -n monitoring
kubectl get crd servicemonitors.monitoring.coreos.com
```

The chart also installs Prometheus, Grafana, Alertmanager and exporters. Once the CRD exists, apply the API `ServiceMonitor`:

```bash
kubectl apply -f k8s/monitoring/short-link-servicemonitor.yaml
kubectl get servicemonitor -n monitoring short-link
```

Prometheus will discover the `ServiceMonitor` and start scraping `short-link:8080/metrics`. Confirm the target in Prometheus or inspect the resources:

```bash
kubectl get servicemonitor -n monitoring
kubectl get prometheus -n monitoring
kubectl get svc -n monitoring
```

To access Prometheus locally:

```bash
kubectl port-forward -n monitoring svc/monitoring-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090` in a browser. To access Grafana, use another terminal:

```bash
kubectl port-forward -n monitoring svc/monitoring-grafana 3000:80
```

Open `http://localhost:3000`. The default credentials can be retrieved from the Secret created by the chart:

```bash
kubectl get secret monitoring-grafana -n monitoring \
  -o jsonpath='{.data.admin-user}' | base64 -d; echo
kubectl get secret monitoring-grafana -n monitoring \
  -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

To upgrade an existing installation, use `helm upgrade` instead of `helm install`:

```bash
helm upgrade monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring
```

## 10. Verification and troubleshooting

```bash
kubectl get all -n short-link
kubectl get events -n short-link --sort-by=.lastTimestamp
kubectl logs deployment/short-link -n short-link
kubectl describe pod -n short-link -l app=short-link
```

Common issues:

- `ImagePullBackOff`: the image was not loaded into the correct cluster; repeat `kind load docker-image`.
- API restarting: check PostgreSQL, Redis and the manual migration.
- `GatewayClass` without controller: install Envoy Gateway.
- HPA without metrics: verify Metrics Server with `kubectl top pods -n short-link`.

## Cleanup

```bash
kubectl delete namespace short-link
kind delete cluster --name short-link
```

Deleting the namespace removes the workloads and PostgreSQL PVCs. The local Docker image remains and can be removed with:

```bash
docker image rm short-link:local
```

## Makefile shortcuts

```bash
make k8s-create
make k8s-build
make k8s-load-image
make k8s-apply-namespace
make k8s-deploy
make k8s-status
make k8s-pods
```

For a full deployment, prefer the ordered commands documented in this guide. The Makefile shortcuts are partial and do not handle PostgreSQL, Redis or the manual migration.
