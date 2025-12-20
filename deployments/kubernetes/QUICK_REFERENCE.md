# Kubernetes Quick Reference

Quick commands for common Gorax deployment operations.

## Deployment Commands

### Deploy to Environment

```bash
# Development
kubectl apply -k deployments/kubernetes/overlays/development

# Staging
kubectl apply -k deployments/kubernetes/overlays/staging

# Production
kubectl apply -k deployments/kubernetes/overlays/production
```

### Preview Changes

```bash
# See what will be deployed
kubectl kustomize deployments/kubernetes/overlays/production

# Show diff before applying
kubectl diff -k deployments/kubernetes/overlays/production
```

## Secrets Management

### Create Development Secrets

```bash
export DB_URL="postgresql://user:password@localhost:5432/gorax"
export JWT_SECRET=$(openssl rand -base64 32)
export ENCRYPTION_KEY=$(openssl rand -base64 32)

kubectl create secret generic dev-gorax-secrets \
  --from-literal=database-url="$DB_URL" \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=encryption-key="$ENCRYPTION_KEY" \
  --namespace=gorax-dev
```

### Verify Secrets

```bash
kubectl get secrets -n gorax-prod
kubectl describe secret prod-gorax-secrets -n gorax-prod
```

## Monitoring

### View Pods

```bash
# All pods
kubectl get pods -n gorax-prod

# Watch pods in real-time
kubectl get pods -n gorax-prod -w

# Wide output (node, IP, etc.)
kubectl get pods -n gorax-prod -o wide

# Filter by component
kubectl get pods -n gorax-prod -l component=api
kubectl get pods -n gorax-prod -l component=worker
```

### View Logs

```bash
# API logs
kubectl logs -f -n gorax-prod -l component=api --tail=100

# Worker logs
kubectl logs -f -n gorax-prod -l component=worker --tail=100

# Specific pod
kubectl logs -f -n gorax-prod pod/prod-gorax-api-xxx

# Previous container (after crash)
kubectl logs -n gorax-prod pod/prod-gorax-api-xxx --previous
```

### Resource Usage

```bash
# Pod resource usage
kubectl top pods -n gorax-prod

# Node resource usage
kubectl top nodes

# Sort by CPU
kubectl top pods -n gorax-prod --sort-by=cpu

# Sort by memory
kubectl top pods -n gorax-prod --sort-by=memory
```

## Scaling

### Manual Scaling

```bash
# Scale API
kubectl scale deployment prod-gorax-api -n gorax-prod --replicas=5

# Scale workers
kubectl scale deployment prod-gorax-worker -n gorax-prod --replicas=10
```

### HPA Status

```bash
# View HPA
kubectl get hpa -n gorax-prod

# Watch HPA
kubectl get hpa -n gorax-prod -w

# Describe HPA
kubectl describe hpa prod-gorax-api-hpa -n gorax-prod
```

## Troubleshooting

### Describe Resources

```bash
# Describe pod
kubectl describe pod -n gorax-prod prod-gorax-api-xxx

# Describe deployment
kubectl describe deployment -n gorax-prod prod-gorax-api

# Describe service
kubectl describe service -n gorax-prod prod-gorax-api
```

### Execute Commands in Pod

```bash
# Get shell
kubectl exec -it -n gorax-prod pod/prod-gorax-api-xxx -- /bin/sh

# Run single command
kubectl exec -it -n gorax-prod pod/prod-gorax-api-xxx -- env

# Test database connection
kubectl exec -it -n gorax-prod pod/prod-gorax-api-xxx -- \
  wget -qO- localhost:8080/health/ready
```

### Port Forwarding

```bash
# Forward API port
kubectl port-forward -n gorax-prod svc/prod-gorax-api 8080:80

# Forward to specific pod
kubectl port-forward -n gorax-prod pod/prod-gorax-api-xxx 8080:8080
```

### Events

```bash
# View recent events
kubectl get events -n gorax-prod --sort-by='.lastTimestamp'

# Watch events
kubectl get events -n gorax-prod -w
```

## Rollout Management

### Update Image

```bash
# Update API
kubectl set image deployment/prod-gorax-api \
  api=gorax/api:v1.2.3 \
  -n gorax-prod

# Update worker
kubectl set image deployment/prod-gorax-worker \
  worker=gorax/worker:v1.2.3 \
  -n gorax-prod
```

### Rollout Status

```bash
# Watch rollout
kubectl rollout status deployment/prod-gorax-api -n gorax-prod

# View history
kubectl rollout history deployment/prod-gorax-api -n gorax-prod

# View specific revision
kubectl rollout history deployment/prod-gorax-api -n gorax-prod --revision=2
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/prod-gorax-api -n gorax-prod

# Rollback to specific revision
kubectl rollout undo deployment/prod-gorax-api -n gorax-prod --to-revision=3
```

### Restart

```bash
# Restart deployment (rolling restart)
kubectl rollout restart deployment/prod-gorax-api -n gorax-prod
kubectl rollout restart deployment/prod-gorax-worker -n gorax-prod
```

## Configuration Updates

### Update ConfigMap

```bash
# Edit ConfigMap
kubectl edit configmap prod-gorax-config -n gorax-prod

# Update from file
kubectl create configmap prod-gorax-config \
  --from-file=config.yaml \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart to pick up changes
kubectl rollout restart deployment/prod-gorax-api -n gorax-prod
```

### Update Secrets

```bash
# Edit secret
kubectl edit secret prod-gorax-secrets -n gorax-prod

# Restart to pick up changes
kubectl rollout restart deployment/prod-gorax-api -n gorax-prod
kubectl rollout restart deployment/prod-gorax-worker -n gorax-prod
```

## Ingress

### View Ingress

```bash
# List ingresses
kubectl get ingress -n gorax-prod

# Describe ingress
kubectl describe ingress prod-gorax-ingress -n gorax-prod

# View ingress YAML
kubectl get ingress prod-gorax-ingress -n gorax-prod -o yaml
```

### Check Certificates

```bash
# View certificates (cert-manager)
kubectl get certificate -n gorax-prod

# Describe certificate
kubectl describe certificate prod-gorax-tls -n gorax-prod

# View certificate secret
kubectl get secret prod-gorax-tls -n gorax-prod -o yaml
```

## Database Operations

### Run Migration Job

```bash
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: gorax-migration-$(date +%s)
  namespace: gorax-prod
spec:
  template:
    spec:
      restartPolicy: Never
      serviceAccountName: prod-gorax
      containers:
      - name: migrate
        image: gorax/api:latest
        command: ["/app/migrate"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: prod-gorax-secrets
              key: database-url
EOF
```

### Check Job Status

```bash
# List jobs
kubectl get jobs -n gorax-prod

# View job logs
kubectl logs -n gorax-prod job/gorax-migration-xxx

# Delete completed job
kubectl delete job -n gorax-prod gorax-migration-xxx
```

## Cleanup

### Delete Resources

```bash
# Delete by label
kubectl delete all -n gorax-prod -l app=gorax

# Delete namespace (removes everything)
kubectl delete namespace gorax-prod

# Delete specific deployment
kubectl delete deployment -n gorax-prod prod-gorax-api
```

### Prune Unused Resources

```bash
# Prune resources not in kustomize
kubectl apply -k deployments/kubernetes/overlays/production --prune
```

## Debugging

### Network Debugging

```bash
# Run debug pod
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -n gorax-prod -- /bin/bash

# Test connectivity to service
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n gorax-prod -- \
  curl http://prod-gorax-api.gorax-prod.svc.cluster.local/health/live

# DNS lookup
kubectl run -it --rm debug --image=busybox --restart=Never -n gorax-prod -- \
  nslookup prod-gorax-api.gorax-prod.svc.cluster.local
```

### Resource Debugging

```bash
# Check resource quotas
kubectl get resourcequota -n gorax-prod

# Check limit ranges
kubectl get limitrange -n gorax-prod

# Check PodDisruptionBudgets
kubectl get pdb -n gorax-prod
```

## Backup and Export

### Export All Resources

```bash
# Export all resources
kubectl get all -n gorax-prod -o yaml > backup-all.yaml

# Export specific resources
kubectl get deployment,service,ingress -n gorax-prod -o yaml > backup-core.yaml

# Export ConfigMaps and Secrets
kubectl get configmap,secret -n gorax-prod -o yaml > backup-config.yaml
```

### Get Resource Manifests

```bash
# Get deployment YAML
kubectl get deployment prod-gorax-api -n gorax-prod -o yaml > api-deployment.yaml

# Get without managed fields
kubectl get deployment prod-gorax-api -n gorax-prod \
  -o yaml --export > api-deployment-clean.yaml
```

## Context and Namespace

### Set Default Namespace

```bash
# Set namespace for current context
kubectl config set-context --current --namespace=gorax-prod

# Verify
kubectl config view --minify | grep namespace
```

### Switch Context

```bash
# List contexts
kubectl config get-contexts

# Switch context
kubectl config use-context production-cluster

# View current context
kubectl config current-context
```

## Useful Aliases

Add these to your `~/.bashrc` or `~/.zshrc`:

```bash
# Kubectl aliases
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgd='kubectl get deployments'
alias kgs='kubectl get services'
alias kl='kubectl logs -f'
alias kdesc='kubectl describe'
alias kex='kubectl exec -it'

# Gorax-specific aliases
alias kgorax='kubectl -n gorax-prod'
alias kgorax-dev='kubectl -n gorax-dev'
alias kgorax-logs-api='kubectl logs -f -n gorax-prod -l component=api'
alias kgorax-logs-worker='kubectl logs -f -n gorax-prod -l component=worker'
```

## Common Scenarios

### Deploy New Version

```bash
# 1. Update image tag in overlay or use kubectl set image
kubectl set image deployment/prod-gorax-api api=gorax/api:v1.2.3 -n gorax-prod

# 2. Watch rollout
kubectl rollout status deployment/prod-gorax-api -n gorax-prod

# 3. Verify
kubectl get pods -n gorax-prod -l component=api
kubectl logs -n gorax-prod -l component=api --tail=50

# 4. Test health
kubectl exec -it -n gorax-prod deploy/prod-gorax-api -- \
  wget -qO- localhost:8080/health/ready
```

### Scale for High Traffic

```bash
# 1. Increase replicas
kubectl scale deployment prod-gorax-api -n gorax-prod --replicas=10
kubectl scale deployment prod-gorax-worker -n gorax-prod --replicas=20

# 2. Monitor
kubectl get hpa -n gorax-prod -w

# 3. Check load
kubectl top pods -n gorax-prod

# 4. Scale back when done
kubectl scale deployment prod-gorax-api -n gorax-prod --replicas=3
kubectl scale deployment prod-gorax-worker -n gorax-prod --replicas=3
```

### Investigate Pod Crash

```bash
# 1. Get pod status
kubectl get pods -n gorax-prod -l component=api

# 2. Check recent events
kubectl get events -n gorax-prod --sort-by='.lastTimestamp' | grep prod-gorax-api

# 3. Describe pod
kubectl describe pod -n gorax-prod <pod-name>

# 4. Check current logs
kubectl logs -n gorax-prod <pod-name>

# 5. Check previous container logs (after crash)
kubectl logs -n gorax-prod <pod-name> --previous

# 6. Get shell if pod is running
kubectl exec -it -n gorax-prod <pod-name> -- /bin/sh
```

## Additional Resources

- Full documentation: `deployments/kubernetes/README.md`
- Kubernetes Cheat Sheet: https://kubernetes.io/docs/reference/kubectl/cheatsheet/
- Kustomize Examples: https://github.com/kubernetes-sigs/kustomize/tree/master/examples
