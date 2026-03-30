package testdb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAcquireAdvisoryLockReportsBlockingSessionIntegration(t *testing.T) {
	holderDB := OpenFromEnv(t)

	holderCtx, holderCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer holderCancel()

	holderConn := MustAcquireAdvisoryLock(t, holderCtx, holderDB, DefaultLockKey)
	if _, err := holderConn.ExecContext(holderCtx, `SET application_name = 'workflow_app_test_lock_holder'`); err != nil {
		t.Fatalf("set holder application name: %v", err)
	}

	waiterDB := OpenFromEnv(t)

	waiterCtx, waiterCancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer waiterCancel()

	_, err := AcquireAdvisoryLock(waiterCtx, waiterDB, DefaultLockKey)
	if err == nil {
		t.Fatal("expected advisory lock acquisition to time out while holder session remains active")
	}

	var timeoutErr *AdvisoryLockTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected AdvisoryLockTimeoutError, got %T: %v", err, err)
	}
	if timeoutErr.LockKey != DefaultLockKey {
		t.Fatalf("expected lock key %d, got %d", DefaultLockKey, timeoutErr.LockKey)
	}
	if timeoutErr.Waited < time.Second {
		t.Fatalf("expected bounded wait of at least 1s, got %s", timeoutErr.Waited)
	}
	if len(timeoutErr.Details) == 0 {
		t.Fatalf("expected blocking-session diagnostics, got none: %v", timeoutErr)
	}

	foundHolder := false
	for _, detail := range timeoutErr.Details {
		if detail.Granted && detail.Application == "workflow_app_test_lock_holder" {
			foundHolder = true
			break
		}
	}
	if !foundHolder {
		t.Fatalf("expected diagnostics to include the holder session, got %+v", timeoutErr.Details)
	}

	if !strings.Contains(timeoutErr.Error(), "workflow_app_test_lock_holder") {
		t.Fatalf("expected timeout message to surface holder diagnostics, got %q", timeoutErr.Error())
	}
}
