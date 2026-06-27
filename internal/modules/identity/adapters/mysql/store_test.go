package mysql

import "testing"

func TestNormalizeBootstrapAdminPasswordRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "empty", password: ""},
		{name: "old default", password: "123456"},
		{name: "too short", password: "1234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeBootstrapAdminPassword(tt.password); err == nil {
				t.Fatal("normalizeBootstrapAdminPassword() error = nil, want validation error")
			}
		})
	}
}

func TestNormalizeBootstrapAdminPasswordAcceptsExplicitPassword(t *testing.T) {
	password, err := normalizeBootstrapAdminPassword("  local-admin-secret  ")
	if err != nil {
		t.Fatalf("normalizeBootstrapAdminPassword() error = %v", err)
	}
	if password != "local-admin-secret" {
		t.Fatalf("password = %q, want local-admin-secret", password)
	}
}
