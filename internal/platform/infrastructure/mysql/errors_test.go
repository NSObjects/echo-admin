package mysql

import (
	"errors"
	"strings"
	"testing"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

func TestMapReadError(t *testing.T) {
	t.Run("record not found maps to not-found naming the resource", func(t *testing.T) {
		got := MapReadError(gorm.ErrRecordNotFound, "role", "list roles")

		if info := apperr.NewInfo(got); info.Code != apperr.ErrNotFound {
			t.Fatalf("code = %d, want %d", info.Code, apperr.ErrNotFound)
		}
		if got.Error() != "role not found" {
			t.Fatalf("message = %q, want %q", got.Error(), "role not found")
		}
	})

	t.Run("other errors stay database failures with operation detail", func(t *testing.T) {
		got := MapReadError(errors.New("connection refused"), "role", "list roles")

		if info := apperr.NewInfo(got); info.Code != apperr.ErrDatabase {
			t.Fatalf("code = %d, want %d", info.Code, apperr.ErrDatabase)
		}
		if got.Error() != "Database error" {
			t.Fatalf("message = %q, want %q", got.Error(), "Database error")
		}
		if !strings.Contains(apperr.NewInfo(got).Detail, "list roles") {
			t.Fatalf("detail = %q, want operation list roles", apperr.NewInfo(got).Detail)
		}
	})
}

func TestMapWriteError(t *testing.T) {
	t.Run("duplicate key 1062 maps to conflict", func(t *testing.T) {
		duplicateKey := &drivermysql.MySQLError{Number: 1062, Message: "Duplicate entry 'root' for key 'idx_code'"}

		got := MapWriteError(duplicateKey, "role code already exists", "create role")

		if info := apperr.NewInfo(got); info.Code != apperr.ErrConflict {
			t.Fatalf("code = %d, want %d", info.Code, apperr.ErrConflict)
		}
		if got.Error() != "role code already exists" {
			t.Fatalf("message = %q, want %q", got.Error(), "role code already exists")
		}
	})

	t.Run("other errors stay database failures", func(t *testing.T) {
		got := MapWriteError(errors.New("deadlock detected"), "role code already exists", "create role")

		if info := apperr.NewInfo(got); info.Code != apperr.ErrDatabase {
			t.Fatalf("code = %d, want %d", info.Code, apperr.ErrDatabase)
		}
	})
}
