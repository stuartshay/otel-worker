# Branch Protection Setup

This document describes the GitHub branch protection rules configured for the otel-worker repository.

## Branch Strategy

- **main**: Protected production branch (requires PR, no direct commits)
- **develop**: Primary development branch (direct commits allowed)
- **feature/\***: Optional feature branches for larger changes

## Current Protection Settings

The `main` branch has the following protections:

✅ **Require a pull request before merging**
- Required approvals: `0` (solo developer)
- Dismiss stale reviews when new commits pushed
- No direct commits to main allowed

✅ **Require status checks to pass before merging**
- Required checks:
  - `Build and Push / build` (docker.yml workflow)
  - `Lint / lint` (lint workflow with golangci-lint)
- Branches must be up to date before merging

✅ **Require conversation resolution before merging**

✅ **Prevent force pushes** - Disabled

✅ **Prevent branch deletion** - Disabled

## Development Workflow

### 1. Work on develop branch

```bash
# Switch to develop
git checkout develop
git pull origin develop

# Make changes to Go code
# ... edit files ...

# Run pre-commit checks
pre-commit run --all-files

# Run tests
go test ./...

# Commit and push
git add .
git commit -m "feat: Add new feature"
git push origin develop
```

### 2. Create Pull Request to main

**Via GitHub CLI**:

```bash
gh pr create --base main --head develop \
  --title "Release: Description" \
  --body "## Changes
- Feature 1
- Fix 2
- Update 3"
```

**Via GitHub Web UI**:

1. https://github.com/stuartshay/otel-worker
2. Pull requests → New pull request
3. Base: `main` ← Compare: `develop`

### 3. CI/CD Pipeline

After creating PR:

1. **Lint workflow runs** - golangci-lint checks code quality
2. **Build workflow runs** - Compiles Go code
3. **Tests run** - Unit tests execute
4. All checks must pass (green ✅)

### 4. Merge Pull Request

**After CI passes**:

```bash
# Merge PR (squash recommended)
gh pr merge <PR_NUMBER> --squash

# Keep develop in sync
git checkout develop
git pull origin main
git push origin develop
```

### 5. Automatic Docker Build

After merge to main:

1. Docker workflow triggers
2. Version: `1.0.<run_number>` (from VERSION file + build #)
3. Multi-arch build (amd64, arm64)
4. Push to `stuartshay/otel-worker:1.0.<run_number>` and `:latest`
5. **repository_dispatch** to k8s-gitops
6. Automated PR created in k8s-gitops
7. Merge k8s-gitops PR → Argo CD deploys

## Updating Dependencies

```bash
# On develop branch
git checkout develop

# Update Go modules
go get -u github.com/lib/pq
go mod tidy

# Verify builds
go build ./...
go test ./...

# Commit and create PR
git add go.mod go.sum
git commit -m "chore(deps): Update lib/pq to v1.11.1"
git push origin develop

gh pr create --base main --head develop \
  --title "chore(deps): Update dependencies"
```

## Checking Branch Protection

```bash
# View current protection settings
gh api repos/stuartshay/otel-worker/branches/main/protection | jq

# Or via web UI
https://github.com/stuartshay/otel-worker/settings/branches
```

## Why Branch Protection?

1. **Prevents accidents** - No force-push or direct commits to main
2. **Ensures quality** - All code passes lint and tests
3. **Maintains history** - Clean, reviewable commit history
4. **Enables automation** - CI/CD triggered by PR merges
5. **Matches patterns** - Consistent with otel-ui and otel-demo

## Emergency: Bypass Protection

If absolutely necessary (production incident):

1. GitHub → Settings → Branches → Edit rule
2. Temporarily disable specific protections
3. Make critical fix
4. **Re-enable immediately**

**Better approach**: Create hotfix PR and merge after fast review.

## Related Documentation

- [](docs/DATABASE.md) - Database schema
- [](docs/DEPLOYMENT_GUIDE.md) - Kubernetes deployment
- [](../k8s-gitops/AGENTS.md) - GitOps workflow
