package mysql

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/platform/configs"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/resources"
)

func TestGORMLoggerIgnoresRecordNotFound(t *testing.T) {
	var output bytes.Buffer
	gormLogger := newGORMLogger(&output, false)

	gormLogger.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT * FROM access_menus WHERE path = '/dashboard'", 0
	}, gorm.ErrRecordNotFound)

	if got := output.String(); got != "" {
		t.Fatalf("logger output = %q, want empty output for record-not-found seed check", got)
	}
}

func TestGORMLoggerReportsDatabaseErrors(t *testing.T) {
	var output bytes.Buffer
	gormLogger := newGORMLogger(&output, false)
	dbErr := errors.New("connection refused")

	gormLogger.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT * FROM access_menus", 0
	}, dbErr)

	got := output.String()
	if !strings.Contains(got, dbErr.Error()) {
		t.Fatalf("logger output = %q, want database error %q", got, dbErr)
	}
}

func TestOpenDisabledMySQLDoesNotConnect(t *testing.T) {
	resource, err := Open(context.Background(), configs.MySQLConfig{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if resource == nil {
		t.Fatal("Open() resource = nil, want disabled resource")
	}
	if resource.DB() != nil {
		t.Fatal("DB() != nil, want disabled resource without DB")
	}

	status := resource.Check(context.Background())
	if status.State != resources.StateDisabled {
		t.Fatalf("State = %q, want %q", status.State, resources.StateDisabled)
	}
}

func TestCheckEnabledMySQLWithoutDBReportsUnavailable(t *testing.T) {
	resource := &Resource{enabled: true}

	status := resource.Check(context.Background())
	if status.Name != resources.CapabilityMySQL {
		t.Fatalf("Name = %q, want %q", status.Name, resources.CapabilityMySQL)
	}
	if !status.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if status.Available {
		t.Fatal("Available = true, want false")
	}
	if status.State != resources.StateUnavailable {
		t.Fatalf("State = %q, want %q", status.State, resources.StateUnavailable)
	}
}

func TestCloseDisabledMySQLIsNoop(t *testing.T) {
	resource := &Resource{}

	if err := resource.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
