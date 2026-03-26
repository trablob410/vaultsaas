# Phase 9: UptimeRobot Monitoring

**Priority:** P1 | **Effort:** 15min | **Status:** pending

Set up uptime monitoring and alerts for production service.

## Steps

1. Go to https://uptimerobot.com (create account if needed)
2. Click "Add Monitor"
3. Configure monitor:
   - **Monitor Type:** HTTP(s)
   - **Friendly Name:** "Valt Production"
   - **URL:** `https://valt.turbo.ai.vn/health`
   - **Check Interval:** 5 minutes
   - **HTTP Method:** GET
4. Add alert contacts:
   - Email: your-email@example.com
   - Slack webhook (optional)
5. Save and activate monitor

## Health Endpoint

Verify health check endpoint exists at `GET /health`:

```bash
curl https://valt.turbo.ai.vn/health
# Expected: {"status":"ok"} with 200 status code
```

If endpoint doesn't exist, add to `server/cmd/server/main.go`:

```go
router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
})
```

## Dashboard Setup (Optional)

In UptimeRobot dashboard:

- [ ] Add public status page: https://uptime.valt.turbo.ai.vn
- [ ] Share link with users so they can check service status
- [ ] Enable incident notifications

## Alerts Configuration

- [ ] Down alert to email within 1 minute
- [ ] Recovery alert when service comes back up
- [ ] Check alert settings before going live

## Verification

- [ ] Monitor shows "Monitoring" status
- [ ] Test alert by stopping server: `docker compose down`
- [ ] Verify alert email arrives within 5 minutes
- [ ] Restart server: `docker compose up -d`
- [ ] Verify recovery email arrives

## Notes

- UptimeRobot free tier: 1 monitor, 5min checks
- Upgrade to PRO for 50+ monitors and 1min checks if needed
- Consider adding database backup monitoring later
