/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/daeuniverse/dae-wing/pkg/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	filename = "wing.db"
)

var (
	db *gorm.DB
)

// InitDatabase opens wing.db, runs schema migrations, and tunes the
// SQLite connection for production load:
//
//   - PRAGMA busy_timeout = 5000
//     SQLite write transactions wait up to 5s for a contended lock
//     instead of failing with SQLITE_BUSY (5) immediately. This was
//     the root cause of the recurring "database is locked" panics
//     after unclean shutdowns (see
//     https://github.com/daeuniverse/daed/issues/205).
//
//   - PRAGMA journal_mode = WAL
//     Readers no longer block writers (and vice versa). wing.db-wal
//     and wing.db-shm appear next to wing.db; the next start
//     automatically re-applies the WAL mode. The only deployment
//     that does not work with WAL is a network filesystem (NFS) —
//     wing.db lives in /etc/daed/ on a local ext4/f2fs, so this
//     is fine.
//
//   - PRAGMA synchronous = NORMAL
//     Crash-safe with WAL (the WAL itself provides the durability
//     guarantee) and skips the per-commit FULL fsync that
//     synchronous=FULL forces. Latency on subscription refreshes
//     drops by an order of magnitude on slow storage.
func InitDatabase(configDir string) (err error) {
	path := filepath.Join(configDir, filename)
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", err, path)
	}
	if err = db.AutoMigrate(
		&User{},
		&Config{},
		&Dns{},
		&Routing{},
		&Node{},
		&Subscription{},
		&Group{},
		&GroupSubscription{},
		&GroupPolicyParam{},
		&System{},
	); err != nil {
		return err
	}

	// Connection-level PRAGMAs. Apply via the underlying *sql.DB so
	// every pooled connection inherits the settings. Run them in a
	// fixed order: busy_timeout first (so the journal_mode change
	// does not block), then journal_mode, then synchronous.
	if err = db.Exec("PRAGMA busy_timeout = 5000").Error; err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}
	if err = db.Exec("PRAGMA journal_mode = WAL").Error; err != nil {
		return fmt.Errorf("set journal_mode: %w", err)
	}
	if err = db.Exec("PRAGMA synchronous = NORMAL").Error; err != nil {
		return fmt.Errorf("set synchronous: %w", err)
	}

	if fi, err := os.Stat(path); err != nil {
		return err
	} else if fi.Mode()&0037 > 0 {
		// Too open, chmod it to 0640.
		if err = os.Chmod(path, 0640); err != nil {
			return err
		}
	}

	return nil
}

func DB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
func SetOutput(writer io.Writer) {
	db.Logger = logger.New(log.New(writer, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})
}
func BeginTx(ctx context.Context) *gorm.DB {
	return DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
}
func BeginReadOnlyTx(ctx context.Context) *gorm.DB {
	return DB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	})
}

// Shutdown gracefully closes the wing.db connection. Callers should
// invoke this from their signal handler (after dae has finished its
// own graceful shutdown) so the SQLite WAL file is checkpointed and
// truncated before the process exits. Idempotent: a second call is
// a no-op.
func Shutdown(ctx context.Context) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("db.DB: %w", err)
	}
	// Best-effort WAL checkpoint. The TRUNCATE mode actively
	// shrinks the -wal file to zero bytes; PASSIVE would just
	// flush. We want TRUNCATE so a subsequent unclean power loss
	// does not leave a multi-megabyte -wal file lying around.
	if err := db.WithContext(ctx).Exec("PRAGMA wal_checkpoint(TRUNCATE)").Error; err != nil {
		logrus.Warnf("db: wal_checkpoint(TRUNCATE) failed: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("sqlDB.Close: %w", err)
	}
	db = nil
	return nil
}
