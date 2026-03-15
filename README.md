# deebee

A standalone database migration CLI written in [Ard](https://ard.run). Manages versioned SQL migrations with up/down pairs, batch tracking, and drift validation.

Supports **PostgreSQL** and **SQLite**.

## Install

Requires the [Ard compiler](https://ard.run).

```bash
# Run directly
ard run main.ard <command>

# Or build a binary
ard build main.ard --out deebee
./deebee <command>
```

## Quick start

```bash
# Initialize the migrations directory
deebee init

# Create a new migration
deebee create add_users

# Edit the generated SQL files
# migrations/0001_add_users.up.sql
# migrations/0001_add_users.down.sql

# Apply all pending migrations
DATABASE_URL=postgres://localhost/mydb deebee up

# Check status
DATABASE_URL=postgres://localhost/mydb deebee status
```

## Commands

| Command | Description |
|---|---|
| `init` | Create the migrations directory |
| `create <name>` | Generate a new up/down migration pair |
| `up` | Apply all pending migrations |
| `down` | Roll back the last batch of migrations |
| `status` | Show applied, pending, and missing migrations |
| `validate` | Validate migration files and check for DB drift |
| `help` | Show usage information |

## Options

| Option | Description | Default |
|---|---|---|
| `--dir <path>` | Migrations directory | `migrations` |
| `--database-url <url>` | Database connection URL | `DATABASE_URL` env var |

## Database URL

The database dialect is inferred from the connection URL:

- **PostgreSQL**: `postgres://user:pass@host:5432/dbname` or `postgresql://...`
- **SQLite**: any other value, typically a file path like `myapp.db`

Set via the `DATABASE_URL` environment variable or pass `--database-url`.

## Migration files

Migrations are numbered SQL file pairs in the migrations directory:

```
migrations/
  0001_create_users.up.sql
  0001_create_users.down.sql
  0002_add_posts.up.sql
  0002_add_posts.down.sql
```

- Filenames follow the pattern `<number>_<name>.<up|down>.sql`
- Numbers are zero-padded to 4 digits
- Every `.up.sql` must have a matching `.down.sql`
- Names must be slugs (no spaces, slashes, or dots)

## How it works

- **Batches**: each `up` run groups all newly applied migrations into a batch. `down` rolls back the most recent batch as a unit.
- **Tracking table**: applied migrations are recorded in a `schema_migrations` table, created automatically on first use.
- **Transactions**: each migration runs in its own transaction. If a migration fails, it is rolled back without affecting others.
- **Ordering**: migrations are applied in order of their numeric prefix.
- **Validation**: `validate` checks that all migration files are properly paired, numbered, and (if a database URL is set) that no applied migrations are missing from disk.

## Project structure

```
ard.toml        # Package config (name = "deebee")
main.ard        # CLI entrypoint
cli.ard         # Argument parsing and usage text
types.ard       # Enums (Command, Dialect, Direction) and structs (Config, Migration)
dialect.ard     # Database dialect inference from URL
files.ard       # Migration file discovery, parsing, creation, and sorting
store.ard       # Migration tracking table operations (schema_migrations)
runner.ard      # Command orchestration (up, down, status, validate)
test/
  cli_test.ard  # Unit tests for CLI parsing
```

## Running tests

```bash
ard test .
```
