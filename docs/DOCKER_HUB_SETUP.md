# Docker Hub Setup Guide

**Last Updated**: 2026-01-25

## Current Status

- ✅ Docker Hub repository exists: `stuartshay/otel-worker`
- ✅ Docker workflow configured with semantic versioning (1.0.{BUILD_NUMBER})
- ✅ Multi-arch builds (amd64/arm64) enabled
- ⚠️ GitHub secrets may need configuration

## Required GitHub Secrets

The Docker workflow automatically pushes images to Docker Hub on merge to `main`.
Ensure these secrets are configured:

| Secret Name | Description | How to Get |
|------------|-------------|------------|
| `DOCKERHUB_USERNAME` | Docker Hub username | Your Docker Hub account name |
| `DOCKERHUB_TOKEN` | Docker Hub access token | Create at <https://hub.docker.com/settings/security> |

## Setup Instructions

### 1. Create Docker Hub Access Token

```bash
# 1. Visit: https://hub.docker.com/settings/security
# 2. Click "New Access Token"
# 3. Description: "otel-worker-github-actions"
# 4. Permissions: Read, Write, Delete
# 5. Copy the token (only shown once)
```

### 2. Add GitHub Secrets

```bash
# Using GitHub CLI
gh secret set DOCKERHUB_USERNAME --body "your-username"
gh secret set DOCKERHUB_TOKEN --body "your-token-here"

# OR via GitHub UI:
# 1. Go to: https://github.com/stuartshay/otel-worker/settings/secrets/actions
# 2. Click "New repository secret"
# 3. Add DOCKERHUB_USERNAME
# 4. Add DOCKERHUB_TOKEN
```

### 3. Verify Workflow

After secrets are set, the next push to `main` will trigger automatic Docker builds:

```bash
# Check if secrets are set (won't show values)
gh secret list

# Trigger a test push
git commit --allow-empty -m "ci: trigger Docker build"
git push origin main

# Monitor workflow
gh run watch

# Expected output: Docker image pushed to docker.io/stuartshay/otel-worker:1.0.{BUILD_NUMBER}
```

## Image Tags

The workflow creates multiple tags for each build:

| Tag Pattern | Example | Purpose |
|------------|---------|---------|
| `1.0.{BUILD_NUMBER}` | `1.0.123` | Semantic version (unique per build) |
| `latest` | `latest` | Always points to most recent main build |
| `main` | `main` | Tracks main branch |
| `sha-{SHORT_SHA}` | `sha-91a0949` | Commit-specific tag |

## Manual Push (If Needed)

If GitHub Actions is unavailable, you can push manually:

```bash
# Build locally
docker build -t stuartshay/otel-worker:1.0.dev .

# Login to Docker Hub
docker login -u your-username

# Push
docker push stuartshay/otel-worker:1.0.dev

# Or use buildx for multi-arch
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag stuartshay/otel-worker:1.0.dev \
  --push \
  .
```

## Verification

Once pushed, verify images are available:

```bash
# Check Docker Hub
curl -s https://hub.docker.com/v2/repositories/stuartshay/otel-worker/tags/ | jq -r '.results[].name'

# Pull and test
docker pull stuartshay/otel-worker:latest
docker run --rm stuartshay/otel-worker:latest --version
```

## Next Steps

After Docker images are pushed:

- ✅ Health checks implemented (blocking issue resolved)
- 🎯 Create Kubernetes manifests (Quick Win #3)
- 🎯 Deploy to k8s-pi5-cluster (Quick Win #5)

See [QUICK_WINS.md](QUICK_WINS.md) for full task list.
