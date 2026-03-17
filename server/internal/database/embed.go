package database

import "embed"

// MigrationsFS embeds all SQL migration files.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
