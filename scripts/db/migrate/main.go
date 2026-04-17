package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	var (
		dsn           = flag.String("dsn", envOrDefault("DB_DSN", "postgres://Beehive-Blog:Beehive-Blog@127.0.0.1:5432/Beehive-Blog?sslmode=disable"), "postgres dsn")
		migrationsDir = flag.String("dir", "sql/migrations", "migrations directory")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}

	if err = ensureSchemaMigrationsTable(ctx, db); err != nil {
		panic(err)
	}

	files, err := listMigrationFiles(*migrationsDir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		version := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		path := filepath.Join(*migrationsDir, f.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		checksum := sha256Hex(sqlBytes)
		applied, err := isApplied(ctx, db, version, checksum)
		if err != nil {
			panic(err)
		}
		if applied {
			fmt.Printf("skip %s\n", version)
			continue
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			panic(err)
		}
		if _, err = tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			panic(fmt.Errorf("apply %s failed: %w", version, err))
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, checksum) VALUES ($1, $2)`, version, checksum); err != nil {
			_ = tx.Rollback()
			panic(fmt.Errorf("record %s failed: %w", version, err))
		}
		if err = tx.Commit(); err != nil {
			panic(err)
		}
		fmt.Printf("applied %s\n", version)
	}

	fmt.Println("migrations completed")
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    checksum VARCHAR(64) NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
	return err
}

func listMigrationFiles(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]os.DirEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") && len(name) >= 8 && name[0] >= '0' && name[0] <= '9' {
			files = append(files, e)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	return files, nil
}

func isApplied(ctx context.Context, db *sql.DB, version, checksum string) (bool, error) {
	var existing string
	err := db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, version).Scan(&existing)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if existing != checksum {
		return false, fmt.Errorf("migration %s checksum mismatch", version)
	}
	return true, nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
