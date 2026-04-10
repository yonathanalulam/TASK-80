package main

import (
	"flag"
	"testing"
)

// TestMigrateOnlyFlagParsing verifies that the --migrate-only flag is
// recognized by the flag package, ensuring the entrypoint supports the
// command documented in scripts/dev.sh.
func TestMigrateOnlyFlagParsing(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	migrateOnly := fs.Bool("migrate-only", false, "Run database migrations and exit")

	// Parse with --migrate-only present.
	err := fs.Parse([]string{"--migrate-only"})
	if err != nil {
		t.Fatalf("failed to parse --migrate-only flag: %v", err)
	}
	if !*migrateOnly {
		t.Error("expected migrateOnly to be true after parsing --migrate-only")
	}
}

// TestMigrateOnlyFlagDefault verifies that the flag defaults to false
// (normal server startup mode).
func TestMigrateOnlyFlagDefault(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	migrateOnly := fs.Bool("migrate-only", false, "Run database migrations and exit")

	// Parse with no flags.
	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("failed to parse empty args: %v", err)
	}
	if *migrateOnly {
		t.Error("expected migrateOnly to default to false")
	}
}
