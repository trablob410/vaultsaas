# Phase 11: Deployment Guide

**Priority:** P2 | **Effort:** 2h | **Status:** pending

Create comprehensive deployment & operations guide for running Valt in production.

## Document: `docs/deployment-guide.md`

### Sections

**1. Pre-Deployment Checklist**
- [ ] VPS provisioned (Ubuntu 20.04+, 2GB RAM, 20GB disk)
- [ ] Docker & Docker Compose installed
- [ ] DNS configured (valt.turbo.ai.vn)
- [ ] SSL certificate ready (Caddy auto-renews)
- [ ] All env vars prepared (see below)

**2. Environment Variables**

Document all required + optional env vars:

```
REQUIRED:
- DATABASE_URL
- MINIO_ACCESS_KEY / MINIO_SECRET_KEY
- JWT_PRIVATE_KEY_PATH / JWT_PUBLIC_KEY_PATH
- VAULT_MASTER_KEY (base64 32-byte key)
- GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET

BILLING:
- STRIPE_SECRET_KEY
- STRIPE_WEBHOOK_SECRET
- STRIPE_PRO_PRICE_ID
- STRIPE_TEAM_PRICE_ID

NOTIFICATIONS:
- SMTP_HOST / SMTP_PORT / SMTP_USER / SMTP_PASSWORD / SMTP_FROM

OPTIONAL:
- REDIS_URL (enables rate limiting)
- SLACK_CLIENT_ID / SLACK_CLIENT_SECRET (Slack OAuth)
- ZALO_OA_TOKEN / ZALO_OA_ID (Zalo notifications)
```

**3. Installation Steps**

```bash
git clone https://github.com/vaultsaas/vault.git /opt/valt
cd /opt/valt
cp .env.production.example .env
# Edit .env with production values
docker compose -f docker-compose.prod.yml up -d
```

**4. Initial Setup**

- [ ] Run migrations: `docker compose exec postgres valt-migrate up`
- [ ] Seed initial data (admin user)
- [ ] Test health endpoint: `curl https://valt.turbo.ai.vn/health`
- [ ] Verify all services running: `docker compose ps`

**5. Backup & Recovery**

```bash
# Daily backup (cron job)
docker compose exec postgres pg_dump -U valt valt > backup.sql

# Restore from backup
docker compose exec postgres psql -U valt < backup.sql

# MinIO backup
aws s3 sync s3://valt-secrets s3://backup-bucket/secrets-backup/
```

**6. Monitoring & Logs**

```bash
# View logs
docker compose logs -f server

# Check uptime
curl https://uptime.valt.turbo.ai.vn

# Database health
docker compose exec postgres psql -U valt -c "SELECT version();"
```

**7. Scaling Considerations**

- Single VPS suitable for <1000 users
- For larger scale: multi-region, managed Postgres, CDN
- Monitor CPU/memory usage
- Plan for 2x traffic growth

**8. Security Checklist**

- [ ] Firewall: Only ports 80, 443, 22 open
- [ ] SSH: Key-based auth only
- [ ] `.env` file: Not version-controlled, mode 600
- [ ] JWT keys: Stored securely, rotated annually
- [ ] Master key: Backed up securely, never in logs
- [ ] Regular updates: `docker pull` + restart monthly

**9. Incident Response**

**Service Down:**
1. Check status page: https://uptime.valt.turbo.ai.vn
2. SSH into VPS: `ssh valt@valt.turbo.ai.vn`
3. Restart services: `docker compose restart`
4. Check logs: `docker compose logs server | tail -100`
5. Escalate if logs show database errors

**Database Down:**
1. Verify Postgres running: `docker compose ps postgres`
2. Check disk space: `df -h`
3. Restart Postgres: `docker compose restart postgres`
4. Restore from backup if corrupted

**Certificate Expired:**
Caddy auto-renews Let's Encrypt certs. If manual renewal needed:
```bash
docker compose exec caddy caddy reload
```

**10. Monitoring Dashboards**

- UptimeRobot: https://uptimerobot.com
- Server logs: `docker compose logs`
- Database: `psql valt`
- MinIO: `http://localhost:9001` (internal only)

**11. Contact & Escalation**

- Alert email: ops@valt.turbo.ai.vn
- On-call: Check rotation in wiki
- Incident slack: #incidents

## Sign-Off

Guide is clear and actionable. New ops person can follow it without questions.
