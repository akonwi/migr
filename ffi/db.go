// Package ffi hosts the Go-side database support for migr's local `sql`
// module. It replaces the removed `ard/sql` stdlib module.
//
// Ard can't drive database/sql directly (variadic ...any parameters,
// pointer-out Scan). This package exposes a small, value-in / value-out
// API that sql.ard wraps.
//
// Pure-Go drivers are used for every supported dialect so a static
// (CGO-free) binary is possible on Alpine.
package ffi

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// DB and Tx are opaque handles the Ard side only passes around.

type DB struct {
	inner *sql.DB
}

type Tx struct {
	inner *sql.Tx
}

// Open connects to a database. Driver names correspond to database/sql
// driver registrations from the blank imports above:
//   - "pgx"    -> github.com/jackc/pgx/v5/stdlib
//   - "sqlite" -> modernc.org/sqlite (pure Go)
//   - "mysql"  -> github.com/go-sql-driver/mysql (pure Go)
func Open(driver, url string) (*DB, error) {
	inner, err := sql.Open(driver, url)
	if err != nil {
		return nil, err
	}
	if err := inner.Ping(); err != nil {
		inner.Close()
		return nil, err
	}
	// SQLite is single-writer; keep the pool small and predictable.
	// For pgx/mysql this is still safe, just conservative.
	if driver == "sqlite" {
		inner.SetMaxOpenConns(1)
	}
	return &DB{inner: inner}, nil
}

func Close(db *DB) error {
	return db.inner.Close()
}

// --- Non-transactional operations ---

func ExecDB(db *DB, query string, args []any) error {
	_, err := db.inner.Exec(query, args...)
	return err
}

func QueryDB(db *DB, query string, args []any) ([]any, error) {
	return scanRows(db.inner.Query(query, args...))
}

func Begin(db *DB) (*Tx, error) {
	tx, err := db.inner.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{inner: tx}, nil
}

// --- Transactional operations ---

func ExecTx(tx *Tx, query string, args []any) error {
	_, err := tx.inner.Exec(query, args...)
	return err
}

func QueryTx(tx *Tx, query string, args []any) ([]any, error) {
	return scanRows(tx.inner.Query(query, args...))
}

func Commit(tx *Tx) error {
	return tx.inner.Commit()
}

func Rollback(tx *Tx) error {
	return tx.inner.Rollback()
}

// --- Internal ---

// scanRows collects every row as an `any` holding a column->value map.
// The outer slice type is []any (not []map[string]any) so it surfaces on
// the Ard side as `[Any]` — each row is an opaque Any that the caller
// decodes with the local `decode` module. []byte columns are converted
// to strings so they arrive as Ard Str.
func scanRows(rows *sql.Rows, err error) ([]any, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var out []any
	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalize(values[i])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// normalize converts values scanned from database/sql into shapes the Ard
// side sees as its native scalar types. Without this, integer columns come
// through as int64 (Ard: Int64, a sized type) rather than the default Int,
// forcing every decoder to unsafe::cast<Int64> instead of the natural Int.
func normalize(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	case int64:
		return int(x)
	case int32:
		return int(x)
	case float32:
		return float64(x)
	default:
		return v
	}
}
