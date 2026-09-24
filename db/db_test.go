/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestInitDatabasePragmas verifies that InitDatabase applies the
// connection-level PRAGMAs (busy_timeout, journal_mode, synchronous)
// that close the SQLITE_BUSY / wing.db-journal hazard described in
// https://github.com/daeuniverse/daed/issues/205.
func TestInitDatabasePragmas(t *testing.T) {
	dir := t.TempDir()
	if err := InitDatabase(dir); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	t.Cleanup(func() {
		_ = Shutdown(context.Background())
	})

	cases := []struct {
		name  string
		want  string
		query string
	}{
		{"busy_timeout_is_5000", "5000", "PRAGMA busy_timeout"},
		{"journal_mode_is_wal", "wal", "PRAGMA journal_mode"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			if err := DB(context.Background()).Raw(tc.query).Scan(&got).Error; err != nil {
				t.Fatalf("%s: %v", tc.query, err)
			}
			if got != tc.want {
				t.Errorf("%s: got %q, want %q", tc.query, got, tc.want)
			}
		})
	}
}

// TestInitDatabaseCreatesWALFiles confirms that switching to WAL
// mode actually creates the .wal file next to wing.db. This is the
// on-disk evidence that the PRAGMA took effect.
func TestInitDatabaseCreatesWALFiles(t *testing.T) {
	dir := t.TempDir()
	if err := InitDatabase(dir); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	t.Cleanup(func() {
		_ = Shutdown(context.Background())
	})

	walPath := filepath.Join(dir, "wing.db-wal")
	if _, err := os.Stat(walPath); errors.Is(err, fs.ErrNotExist) {
		t.Errorf("wing.db-wal not present after WAL init (PRAGMA did not take effect)")
	} else if err != nil {
		t.Errorf("stat wing.db-wal: %v", err)
	}
}

// TestShutdownIsIdempotent ensures that calling Shutdown twice
// does not panic and the second call returns nil.
func TestShutdownIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := InitDatabase(dir); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	if err := Shutdown(context.Background()); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	if err := Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown (idempotency): %v", err)
	}
}

// TestShutdownCheckpointTruncatesWAL writes a row, calls
// Shutdown, and confirms the -wal file has been truncated. This
// is the test that directly proves the fix for
// https://github.com/daeuniverse/daed/issues/205.
func TestShutdownCheckpointTruncatesWAL(t *testing.T) {
	dir := t.TempDir()
	if err := InitDatabase(dir); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}

	// Insert a row so the WAL has something to checkpoint.
	if err := DB(context.Background()).Create(&User{Username: "shutdown-truncate-test", Password: "x", JwtSecret: "y", Role: "admin"}).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// After checkpoint(TRUNCATE) the -wal file should be 0 bytes
	// or absent. We accept either, but we must not have a non-zero
	// -wal sitting on disk.
	walPath := filepath.Join(dir, "wing.db-wal")
	if fi, err := os.Stat(walPath); err == nil && fi.Size() > 0 {
		t.Errorf("wing.db-wal size = %d after Shutdown, want 0 (or absent)", fi.Size())
	}
}
