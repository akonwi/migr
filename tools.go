//go:build tools

// This file pins Go module dependencies referenced only by the
// Ard-generated code (chi-style), so that `go mod tidy` retains them
// and records their full go.sum closure. Excluded from normal builds
// by the `tools` build tag.
package tools

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)
