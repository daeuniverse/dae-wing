//go:build mips || mipsle || mips64le || mips64

// mips archtectures are not supported by pure go implementation "github.com/glebarez/sqlite".
// NOTE: CGO_ENABLED=1 is requisite in runtime.

package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}
