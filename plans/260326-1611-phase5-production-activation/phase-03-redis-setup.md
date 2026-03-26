# Phase 3: Redis for Rate Limiting

**Priority:** P1 | **Effort:** 30min | **Status:** pending

Enable Redis for agent rate limiting (60 rpm default).

## Steps

1. SSH into VPS
2. Edit `docker-compose.prod.yml`, add Redis service:

```yaml
redis:
  image: redis:7-alpine
  container_name: valt-redis
  ports:
    - "6379:6379"
  volumes:
    - valt-redis-data:/data
  restart: unless-stopped

volumes:
  valt-redis-data:
```

3. Edit `.env`:

```bash
REDIS_URL=redis://redis:6379
```

4. Restart stack:

```bash
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

## Verification

1. Check Redis is running: `docker compose logs redis`
2. Test connection: `redis-cli -h redis ping` (should return PONG)
3. Make API requests with `X-Agent-ID` header, verify rate limit headers:
   - `X-RateLimit-Limit: 60`
   - `X-RateLimit-Remaining: 59`
4. Hit endpoint 60 times quickly, verify 429 (Too Many Requests)

## Notes

- Persists rate limit state across restarts
- Optional: Set `REDIS_URL` env var for HA Redis (e.g., Redis Cloud)
