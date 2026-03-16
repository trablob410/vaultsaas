# Valt - AI Secret Vault

MCP-native secret vault with human-in-the-loop approval workflow for AI agents.

## What is Valt?

Valt lets developers store secrets securely and control AI agent access through an approval workflow:

1. Store secrets encrypted end-to-end (zero-knowledge)
2. AI agents request access via MCP Protocol
3. You approve/reject from the dashboard
4. AI gets temporary, auto-expiring credentials
5. Everything is audit-logged

## Architecture

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Backend | Go 1.22+ | Auth, vault, workflow, audit API |
| Dashboard | Next.js 15 | Web UI for secret & approval management |
| MCP Server | Rust | Local MCP server for AI agent integration |
| Database | PostgreSQL 16 | Metadata, audit logs |
| Object Store | MinIO | Encrypted blob storage |
| Proxy | Caddy | Reverse proxy + auto TLS |

## Quick Start (Development)

```bash
# Clone and configure
git clone https://github.com/your-org/valt.git
cd valt
cp .env.example .env

# Start all services
docker compose up -d

# Run migrations
make migrate-up

# Seed dev data
make seed

# Access:
# API:       http://localhost:8080
# Dashboard: http://localhost:3000
# MinIO:     http://localhost:9001
```

## Project Structure

```
valt/
├── server/          # Go backend monolith
├── dashboard/       # Next.js web dashboard
├── mcp-server/      # Rust MCP server
├── docs/            # Documentation
├── scripts/         # Dev/ops scripts
└── docker-compose.yml
```

## Development

```bash
make help          # Show all commands
make dev           # Start dev environment
make test          # Run all tests
make lint          # Run linters
make build         # Build all services
```

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting and security model details.

## License

Apache License 2.0 - see [LICENSE](LICENSE).
