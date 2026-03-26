---
phase: "4.8"
title: "Monitoring + Backups"
priority: P1
effort: 2h
status: pending
---

# Phase 4.8: Monitoring + Backups

## Context Links
- `docker-compose.prod.yml` -- production services
- `server/cmd/server/main.go` -- server entry point (no /health endpoint visible)

## Overview

No monitoring or backup strategy. Need basic uptime monitoring and automated Postgres backups.

## Requirements

### Functional
- Health endpoint: `GET /health` returns 200 + JSON status
- UptimeRobot monitors /health endpoint every 5 minutes
- Daily Postgres backup with 7-day retention
- Backup restore documented

### Non-functional
- Backup must not impact production performance significantly
- Backup files stored on VPS local disk (sufficient for single-VPS)
- Alert on downtime via email/Telegram

## Implementation Steps

### Step 1: Health endpoint

Add to `server/cmd/server/main.go` or a dedicated handler:
```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check DB connection
    err := pool.Ping(r.Context())
    status := "ok"
    if err != nil { status = "degraded" }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": status,
        "version": "1.0.0",
    })
})
```

Mount at root level (not under `/api/v1/`) so it's publicly accessible without auth.

### Step 2: UptimeRobot setup

1. Create free account at uptimerobot.com
2. Add HTTP monitor: `https://valt.turbo.ai.vn/health`
3. Interval: 5 minutes
4. Alert contacts: email + Telegram (if bot available)

### Step 3: Postgres backup cron

Option A: Host cron job (simpler, no Docker changes):
```bash
# /etc/cron.d/valt-backup
0 3 * * * root docker exec valt-postgres pg_dump -U $POSTGRES_USER $POSTGRES_DB | gzip > /opt/valt/backups/valt-$(date +\%Y\%m\%d).sql.gz 2>&1
```

Option B: Dedicated backup service in docker-compose (self-contained):
```yaml
# Add to docker-compose.prod.yml
backup:
  image: postgres:16-alpine
  entrypoint: /bin/sh
  command: >
    -c 'while true; do
      pg_dump -h postgres -U $$POSTGRES_USER $$POSTGRES_DB | gzip > /backups/valt-$$(date +%Y%m%d-%H%M).sql.gz;
      find /backups -name "*.sql.gz" -mtime +7 -delete;
      sleep 86400;
    done'
  environment:
    POSTGRES_USER: ${POSTGRES_USER}
    PGPASSWORD: ${POSTGRES_PASSWORD}
    POSTGRES_DB: ${POSTGRES_DB}
  volumes:
    - /opt/valt/backups:/backups
  depends_on:
    postgres:
      condition: service_healthy
  networks:
    - backend
  restart: unless-stopped
```

**Recommendation:** Option B (Docker service) -- self-contained, survives host cron config loss.

### Step 4: Backup retention
7-day retention via `find /backups -name "*.sql.gz" -mtime +7 -delete` in the backup loop.

### Step 5: Restore documentation

Add to `docs/deployment-guide.md`:
```markdown
## Restore from Backup
1. Stop server: `docker compose stop server dashboard`
2. Restore: `gunzip -c /opt/valt/backups/valt-YYYYMMDD.sql.gz | docker exec -i valt-postgres psql -U $POSTGRES_USER $POSTGRES_DB`
3. Start server: `docker compose start server dashboard`
```

### Step 6: Create backup directory on VPS
```bash
mkdir -p /opt/valt/backups
```

## Todo Checklist

- [ ] Add `GET /health` endpoint (DB ping + status)
- [ ] Set up UptimeRobot monitor
- [ ] Add backup service to docker-compose.prod.yml
- [ ] Create /opt/valt/backups directory on VPS
- [ ] Test backup: verify .sql.gz file created
- [ ] Test restore: restore backup to fresh DB
- [ ] Document restore procedure in deployment guide
- [ ] Configure UptimeRobot alerts (email)

## Success Criteria

- /health returns 200 with DB status
- UptimeRobot monitoring active, alerts on downtime
- Daily backups created in /opt/valt/backups/
- 7-day retention working
- Restore procedure documented and tested

## Security Considerations

- Backup files contain unencrypted DB data -- restrict permissions (chmod 600)
- /health endpoint should not expose sensitive info (version only, no config)
- Consider encrypting backups if VPS is shared (not needed for dedicated VPS)
