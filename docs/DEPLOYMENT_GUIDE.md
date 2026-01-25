# Deployment Guide - otel-worker

**Last Updated**: 2026-01-25
**Status**: Ready for deployment to k8s-pi5-cluster

## Prerequisites Checklist

- ✅ Health check endpoints implemented (/healthz, /readyz)
- ✅ Kubernetes manifests created and validated
- ✅ Docker image built (stuartshay/otel-worker:1.0.1)
- ⚠️ Docker image needs to be pushed to Docker Hub
- ⚠️ Sealed secret needs to be generated
- ⚠️ Argo CD application not configured yet

## Deployment Steps

### Step 1: Configure Docker Hub Secrets (5 minutes)

Required for pushing Docker images from GitHub Actions.

```bash
# Set secrets in GitHub
gh secret set DOCKERHUB_USERNAME --body "stuartshay"
gh secret set DOCKERHUB_TOKEN --body "your-docker-hub-token"

# Verify secrets are set (won't show values)
gh secret list | grep DOCKERHUB

# Trigger Docker build (next push to main will auto-build)
# OR trigger manually:
git commit --allow-empty -m "ci: trigger Docker build"
git push origin main
```

**Verification**:

```bash
# Monitor workflow
gh run list --workflow=docker.yml --limit 1

# After success, verify image exists
docker pull stuartshay/otel-worker:latest
docker images | grep otel-worker
```

---

### Step 2: Generate Sealed Secret (5 minutes)

Required for database credentials in Kubernetes.

```bash
# Navigate to k8s-gitops repo
cd /home/ubuntu/git/k8s-gitops/apps/base/otel-worker

# Create temporary plain secret (DO NOT COMMIT THIS)
kubectl create secret generic postgres-credentials \
  --from-literal=POSTGRES_HOST=192.168.1.175 \
  --from-literal=POSTGRES_PORT=6432 \
  --from-literal=POSTGRES_DB=owntracks \
  --from-literal=POSTGRES_USER=development \
  --from-literal=POSTGRES_PASSWORD=development \
  --namespace=otel-worker \
  --dry-run=client -o yaml > /tmp/postgres-secret.yaml

# Seal the secret using kubeseal
kubeseal --format yaml \
  < /tmp/postgres-secret.yaml \
  > postgres-sealed-secret.yaml

# IMPORTANT: Securely delete temporary file
rm /tmp/postgres-secret.yaml

# Commit sealed secret
cd /home/ubuntu/git/k8s-gitops
git add apps/base/otel-worker/postgres-sealed-secret.yaml
git commit -m "chore: add sealed postgres credentials for otel-worker"
git push origin master
```

---

### Step 3: Create Argo CD Application (Optional - If Using GitOps)

If using Argo CD for automated deployments:

```bash
# Create Argo CD app manifest
cat <<EOF > /tmp/otel-worker-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: otel-worker
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/stuartshay/k8s-gitops
    targetRevision: HEAD
    path: apps/base/otel-worker
  destination:
    server: https://kubernetes.default.svc
    namespace: otel-worker
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
EOF

# Apply to cluster
kubectl apply -f /tmp/otel-worker-app.yaml

# Watch sync status
argocd app get otel-worker
```

---

### Step 4: Manual Deployment (Alternative to Argo CD)

If deploying manually without Argo CD:

```bash
# Ensure cluster context is correct
kubectl config use-context k8s-pi5-cluster

# Validate manifests
kubectl kustomize apps/base/otel-worker/ | kubeconform -summary -

# Apply manifests
kubectl apply -k apps/base/otel-worker/

# Watch deployment rollout
kubectl rollout status deployment/otel-worker -n otel-worker

# Expected output:
# Waiting for deployment "otel-worker" rollout to finish: 0 of 2 updated replicas are available...
# Waiting for deployment "otel-worker" rollout to finish: 1 of 2 updated replicas are available...
# deployment "otel-worker" successfully rolled out
```

---

### Step 5: Verify Deployment (10 minutes)

```bash
# Check all resources
kubectl get all -n otel-worker

# Expected output:
# NAME                               READY   STATUS    RESTARTS   AGE
# pod/otel-worker-xxxxxxxxxx-xxxxx   1/1     Running   0          2m
# pod/otel-worker-xxxxxxxxxx-xxxxx   1/1     Running   0          2m
#
# NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)               AGE
# service/otel-worker   ClusterIP   10.96.xxx.xxx   <none>        50051/TCP,8080/TCP    2m
#
# NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/otel-worker   2/2     2            2           2m

# Check PVC
kubectl get pvc -n otel-worker
# Expected: otel-worker-data should be Bound

# Check logs
kubectl logs -n otel-worker -l app.kubernetes.io/name=otel-worker --tail=50

# Expected log entries:
# {"level":"info","time":"...","message":"Starting otel-worker service"}
# {"level":"info","time":"...","message":"Database connection established"}
# {"level":"info","time":"...","message":"Database health check passed"}
# {"level":"info","time":"...","message":"gRPC server listening","port":"50051"}
# {"level":"info","time":"...","message":"HTTP health server listening","port":"8080"}
```

---

### Step 6: Test Health Endpoints (5 minutes)

```bash
# Port-forward HTTP port
kubectl port-forward -n otel-worker svc/otel-worker 8080:8080 &

# Test liveness probe
curl http://localhost:8080/healthz
# Expected: {"status":"healthy","service":"otel-worker"}

# Test readiness probe
curl http://localhost:8080/readyz
# Expected: {"status":"ready","service":"otel-worker","database":"connected"}

# Clean up port-forward
pkill -f "port-forward.*otel-worker"
```

---

### Step 7: Test gRPC Service (10 minutes)

```bash
# Port-forward gRPC port
kubectl port-forward -n otel-worker svc/otel-worker 50051:50051 &

# List available services (requires grpcurl)
grpcurl -plaintext localhost:50051 list

# Expected output:
# distance.v1.DistanceService
# grpc.health.v1.Health
# grpc.reflection.v1alpha.ServerReflection

# Test ListJobs endpoint
grpcurl -plaintext localhost:50051 distance.v1.DistanceService/ListJobs

# Expected: {"jobs":[]} (empty initially)

# Test job submission (replace with valid date)
grpcurl -plaintext -d '{"date":"2026-01-24","device_id":""}' \
  localhost:50051 \
  distance.v1.DistanceService/CalculateDistanceFromHome

# Expected: {"job_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}

# Check job status
grpcurl -plaintext -d '{"job_id":"<job-id-from-above>"}' \
  localhost:50051 \
  distance.v1.DistanceService/GetJobStatus

# Clean up
pkill -f "port-forward.*otel-worker"
```

---

### Step 8: Verify CSV File Generation (5 minutes)

```bash
# Exec into pod to check CSV files
kubectl exec -n otel-worker deployment/otel-worker -- ls -la /data/

# Expected: CSV files should appear when jobs complete
# Example: distance_20260124.csv

# Check CSV content
kubectl exec -n otel-worker deployment/otel-worker -- cat /data/distance_20260124.csv

# Verify file permissions
kubectl exec -n otel-worker deployment/otel-worker -- stat /data/
```

---

## Troubleshooting

### Pod Not Starting

```bash
# Describe pod for events
kubectl describe pod -n otel-worker -l app.kubernetes.io/name=otel-worker

# Common issues:
# 1. ImagePullBackOff - Docker image not pushed or wrong tag
#    Solution: Verify Docker Hub push, check deployment.yaml image tag
#
# 2. CreateContainerConfigError - Secret not found
#    Solution: Verify sealed secret is created and unsealed
#
# 3. CrashLoopBackOff - App crashing on startup
#    Solution: Check logs for errors, verify database connectivity
```

### Readiness Probe Failing

```bash
# Check readiness probe
kubectl exec -n otel-worker deployment/otel-worker -- \
  wget -qO- http://localhost:8080/readyz

# If database unreachable:
ssh vagrant@192.168.1.175 "pg_isready -h localhost -p 6432"

# Check network policies
kubectl get networkpolicies -n otel-worker

# Verify secret contains correct values
kubectl get secret postgres-credentials -n otel-worker -o yaml
```

### PVC Not Binding

```bash
# Check PVC status
kubectl get pvc -n otel-worker

# If stuck in Pending:
kubectl describe pvc otel-worker-data -n otel-worker

# Verify storage class exists
kubectl get storageclass nfs-general

# If storage class missing, update pvc.yaml to use available class
kubectl get storageclass
```

---

## Rollback Procedure

If deployment fails and you need to rollback:

```bash
# Delete deployment
kubectl delete -k apps/base/otel-worker/

# Or rollback to previous version
kubectl rollout undo deployment/otel-worker -n otel-worker

# Check status
kubectl rollout status deployment/otel-worker -n otel-worker
```

---

## Monitoring After Deployment

### First 24 Hours

```bash
# Watch logs continuously
kubectl logs -n otel-worker -l app.kubernetes.io/name=otel-worker -f

# Check for errors
kubectl logs -n otel-worker -l app.kubernetes.io/name=otel-worker \
  | grep -i error

# Monitor resource usage
kubectl top pods -n otel-worker

# Check restarts
kubectl get pods -n otel-worker -w
```

### Health Checks

```bash
# Automated checks every 5 minutes (add to cron or monitoring)
while true; do
  kubectl exec -n otel-worker deployment/otel-worker -- \
    wget -qO- http://localhost:8080/healthz
  sleep 300
done
```

---

## Next Steps After Deployment

Once successfully deployed:

1. **Verify end-to-end functionality**:
   - Submit a distance calculation job
   - Wait for completion
   - Download and verify CSV file

2. **Run integration tests** (Quick Win #2):
   - Test handler with real database queries
   - Verify CSV generation
   - Test concurrent job processing

3. **Add OpenTelemetry instrumentation** (Quick Win #6):
   - Export traces to New Relic
   - Monitor gRPC calls
   - Track database queries

4. **Integrate with otel-demo API** (Quick Win #8):
   - Add REST endpoint wrapper
   - Enable frontend access

See [QUICK_WINS.md](QUICK_WINS.md) for full roadmap.

---

## Production Readiness Checklist

Before considering production-ready:

- [ ] Sealed secret generated and applied
- [ ] Docker images pushed to Docker Hub
- [ ] Deployed to k8s-pi5-cluster
- [ ] Health probes passing
- [ ] gRPC service responding
- [ ] CSV files generated successfully
- [ ] No errors in logs for 24 hours
- [ ] Integration tests passing (handler coverage > 50%)
- [ ] OpenTelemetry traces visible
- [ ] Resource usage within limits (< 256Mi memory, < 200m CPU)

---

## Support

**Issues**: <https://github.com/stuartshay/otel-worker/issues>
**Related Projects**:

- [k8s-gitops](https://github.com/stuartshay/k8s-gitops) - Kubernetes manifests
- [otel-demo](https://github.com/stuartshay/otel-demo) - REST API integration
- [otel-ui](https://github.com/stuartshay/otel-ui) - Frontend UI
