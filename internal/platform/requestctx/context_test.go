package requestctx

import (
	"context"
	"strings"
	"testing"
)

const (
	requestIDForTest = "req-789"
	traceIDForTest   = "trace-123"
	userIDForTest    = "user-001"
	roleIDForTest    = "role-001"
	sessionIDForTest = "session-001"
)

func TestWithInfoStoresRequestMetadata(t *testing.T) {
	ctx := WithInfo(context.Background(), Info{
		TraceID:        traceIDForTest,
		SpanID:         "span-456",
		RequestID:      requestIDForTest,
		UserID:         userIDForTest,
		RoleID:         roleIDForTest,
		LoginSessionID: sessionIDForTest,
	})

	info, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}
	if info.TraceID != traceIDForTest {
		t.Fatalf("TraceID = %q, want %s", info.TraceID, traceIDForTest)
	}
	if info.SpanID != "span-456" {
		t.Fatalf("SpanID = %q, want span-456", info.SpanID)
	}
	if info.RequestID != requestIDForTest {
		t.Fatalf("RequestID = %q, want %s", info.RequestID, requestIDForTest)
	}
	if info.UserID != userIDForTest {
		t.Fatalf("UserID = %q, want %s", info.UserID, userIDForTest)
	}
	if info.RoleID != roleIDForTest {
		t.Fatalf("RoleID = %q, want %s", info.RoleID, roleIDForTest)
	}
	if info.LoginSessionID != sessionIDForTest {
		t.Fatalf("LoginSessionID = %q, want %s", info.LoginSessionID, sessionIDForTest)
	}
}

func TestFromContextWithoutMetadata(t *testing.T) {
	info, ok := FromContext(context.Background())
	if ok {
		t.Fatal("FromContext() ok = true, want false")
	}
	if info != (Info{}) {
		t.Fatalf("Info = %+v, want zero value", info)
	}
}

func TestWithTraceSpanPreservesExistingMetadata(t *testing.T) {
	ctx := WithInfo(context.Background(), Info{
		RequestID:      requestIDForTest,
		UserID:         userIDForTest,
		LoginSessionID: sessionIDForTest,
	})

	ctx = WithTraceSpan(ctx, traceIDForTest, "span-456")

	info, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}
	if info.RequestID != requestIDForTest {
		t.Fatalf("RequestID = %q, want %s", info.RequestID, requestIDForTest)
	}
	if info.UserID != userIDForTest {
		t.Fatalf("UserID = %q, want %s", info.UserID, userIDForTest)
	}
	if info.LoginSessionID != sessionIDForTest {
		t.Fatalf("LoginSessionID = %q, want %s", info.LoginSessionID, sessionIDForTest)
	}
	if info.TraceID != traceIDForTest {
		t.Fatalf("TraceID = %q, want %s", info.TraceID, traceIDForTest)
	}
	if info.SpanID != "span-456" {
		t.Fatalf("SpanID = %q, want span-456", info.SpanID)
	}
}

func TestCleanMetadataID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "visible ascii", value: "req-123", want: "req-123"},
		{name: "space is rejected", value: "req 123", want: ""},
		{name: "non ascii is rejected", value: "请求", want: ""},
		{name: "overlong is rejected", value: strings.Repeat("a", 129), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanMetadataID(tt.value); got != tt.want {
				t.Fatalf("CleanMetadataID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithUserIDAddsAuthenticatedIdentity(t *testing.T) {
	ctx := WithInfo(context.Background(), Info{
		TraceID:   traceIDForTest,
		RequestID: "req-789",
	})

	ctx = WithUserID(ctx, "user-001")

	info, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}
	if info.TraceID != traceIDForTest {
		t.Fatalf("TraceID = %q, want %s", info.TraceID, traceIDForTest)
	}
	if info.UserID != "user-001" {
		t.Fatalf("UserID = %q, want user-001", info.UserID)
	}
}

func TestWithRoleIDAddsActiveAuthenticatedRole(t *testing.T) {
	ctx := WithInfo(context.Background(), Info{
		TraceID:   traceIDForTest,
		RequestID: requestIDForTest,
		UserID:    userIDForTest,
	})

	ctx = WithRoleID(ctx, roleIDForTest)

	info, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}
	if info.UserID != userIDForTest {
		t.Fatalf("UserID = %q, want %s", info.UserID, userIDForTest)
	}
	if info.RoleID != roleIDForTest {
		t.Fatalf("RoleID = %q, want %s", info.RoleID, roleIDForTest)
	}
	if got := GetRoleID(ctx); got != roleIDForTest {
		t.Fatalf("GetRoleID() = %q, want %s", got, roleIDForTest)
	}
}

func TestWithLoginSessionIDAddsSessionMetadata(t *testing.T) {
	ctx := WithUserID(context.Background(), userIDForTest)
	ctx = WithLoginSessionID(ctx, sessionIDForTest)

	info, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}
	if info.UserID != userIDForTest {
		t.Fatalf("UserID = %q, want %s", info.UserID, userIDForTest)
	}
	if got := GetLoginSessionID(ctx); got != sessionIDForTest {
		t.Fatalf("GetLoginSessionID() = %q, want %s", got, sessionIDForTest)
	}
}

func TestContextAccessorsHandleNilContext(t *testing.T) {
	var ctx context.Context
	if got := GetRequestID(ctx); got != "" {
		t.Fatalf("GetRequestID(nil) = %q, want empty", got)
	}
	if got := GetUserID(ctx); got != "" {
		t.Fatalf("GetUserID(nil) = %q, want empty", got)
	}
	if got := GetRoleID(ctx); got != "" {
		t.Fatalf("GetRoleID(nil) = %q, want empty", got)
	}
	if got := GetLoginSessionID(ctx); got != "" {
		t.Fatalf("GetLoginSessionID(nil) = %q, want empty", got)
	}
}
