// Package migrations embeds the SQL migration files into the Go binary.
package migrations

import "embed"

// FS contains every *.sql file in this folder.
//
//go:embed *.sql
var FS embed.FS
