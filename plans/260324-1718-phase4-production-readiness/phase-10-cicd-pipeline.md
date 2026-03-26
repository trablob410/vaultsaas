---
phase: "4.10"
title: "CI/CD Pipeline"
priority: P2
effort: 4h
status: pending
---

# Phase 4.10: CI/CD Pipeline

## Context Links
- `docker-compose.prod.yml` -- production Docker Compose
- `server/Dockerfile` -- Go server Dockerfile
- `dashboard/Dockerfile` -- Next.js Dockerfile
- `Makefile` -- existing build/test commands

## Overview

No automated deployment. Currently manual: SSH to VPS, git pull, docker compose build, docker compose up. Need GitHub Actions for automated lint, test, build, deploy.

## Requirements

### Functional
- On push to master: lint -> test -> build Docker images -> push to GHCR -> deploy to VPS
- On PR: lint -> test only (no deploy)
- Rollback: re-deploy previous image tag
- Slack/email notification on deploy failure

### Non-functional
- Total CI time < 10 minutes
- Zero-downtime deploy (docker compose pull + up -d)
- Secrets stored in GitHub repo secrets

## Implementation Steps

### Step 1: GitHub Actions workflow file

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD

on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

env:
  REGISTRY: ghcr.io
  SERVER_IMAGE: ghcr.io/${{ github.repository }}/server
  DASHBOARD_IMAGE: ghcr.io/${{ github.repository }}/dashboard

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd server && golangci-lint run ./...
      - run: cd dashboard && npm ci && npm run lint

  test:
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: valt_test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: cd server && go test ./internal/... ./pkg/... -v
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/valt_test?sslmode=disable
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd dashboard && npm ci && npm test

  build-and-push:
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/master' && github.event_name == 'push'
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./server
          push: true
          tags: |
            ${{ env.SERVER_IMAGE }}:latest
            ${{ env.SERVER_IMAGE }}:${{ github.sha }}
      - uses: docker/build-push-action@v5
        with:
          context: ./dashboard
          push: true
          tags: |
            ${{ env.DASHBOARD_IMAGE }}:latest
            ${{ env.DASHBOARD_IMAGE }}:${{ github.sha }}

  deploy:
    runs-on: ubuntu-latest
    needs: build-and-push
    if: github.ref == 'refs/heads/master' && github.event_name == 'push'
    steps:
      - uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd /opt/valt
            docker compose pull server dashboard
            docker compose up -d server dashboard
            docker image prune -f
```

### Step 2: GitHub Repository Secrets

Configure in GitHub repo Settings > Secrets:
- `VPS_HOST` -- VPS IP address
- `VPS_USER` -- SSH user (e.g., root or deploy)
- `VPS_SSH_KEY` -- SSH private key for VPS access

### Step 3: Update docker-compose.prod.yml for GHCR images

Replace `build:` directives with `image:` for server and dashboard:
```yaml
server:
  image: ghcr.io/org/valt/server:latest
  # ... rest of config unchanged

dashboard:
  image: ghcr.io/org/valt/dashboard:latest
  # ... rest unchanged
```

Keep build directives as comments for local development.

### Step 4: Rollback procedure

To rollback to a previous version:
```bash
# On VPS
docker compose pull server dashboard  # pulls :latest
# Or specify a commit SHA tag:
# Edit docker-compose to use :abc123 tag
docker compose up -d server dashboard
```

Document in deployment guide.

### Step 5: Deploy notification (optional)

Add a final step to the deploy job:
```yaml
- name: Notify
  if: failure()
  run: |
    curl -X POST "${{ secrets.SLACK_WEBHOOK_URL }}" \
      -H 'Content-Type: application/json' \
      -d '{"text":"Deploy failed: ${{ github.sha }}"}'
```

## Todo Checklist

- [ ] Create `.github/workflows/ci-cd.yml`
- [ ] Configure GitHub repo secrets (VPS_HOST, VPS_USER, VPS_SSH_KEY)
- [ ] Generate deploy SSH key pair, add public key to VPS authorized_keys
- [ ] Update docker-compose.prod.yml for GHCR images
- [ ] Test PR workflow (lint + test only)
- [ ] Test push-to-master workflow (full pipeline)
- [ ] Verify zero-downtime deploy (docker compose up -d)
- [ ] Test rollback procedure
- [ ] Document CI/CD in deployment guide
- [ ] Add deploy failure notification (optional)

## Success Criteria

- PRs automatically linted and tested
- Push to master triggers full deploy pipeline
- Docker images stored in GHCR with commit SHA tags
- VPS updated via SSH with zero downtime
- Rollback possible via image tag

## Security Considerations

- SSH key stored as GitHub secret (not committed)
- Deploy key should have minimal permissions (not root if possible)
- GHCR packages should be private
- GitHub Actions uses GITHUB_TOKEN for GHCR auth (no extra secrets)
- Consider IP allowlisting for SSH access to VPS
