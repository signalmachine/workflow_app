package testdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	DefaultLockKey      int64         = 20260319
	DefaultSetupTimeout time.Duration = 2 * time.Minute

	lockRetryInterval time.Duration = 250 * time.Millisecond
)

type AdvisoryLockSession struct {
	Granted        bool
	PID            int
	Database       string
	User           string
	Application    string
	State          string
	WaitEventType  string
	WaitEvent      string
	TransactionAge string
	QueryAge       string
	Query          string
}

type AdvisoryLockTimeoutError struct {
	LockKey int64
	Waited  time.Duration
	Details []AdvisoryLockSession
	Cause   error
}

func (e *AdvisoryLockTimeoutError) Error() string {
	if e == nil {
		return ""
	}

	message := fmt.Sprintf(
		"timed out after %s waiting for disposable test database advisory lock %d",
		e.Waited.Round(time.Millisecond),
		e.LockKey,
	)
	if e.Cause != nil && !errors.Is(e.Cause, context.DeadlineExceeded) && !errors.Is(e.Cause, context.Canceled) {
		message += fmt.Sprintf(": %v", e.Cause)
	}
	if len(e.Details) == 0 {
		return message + "; no active lock-holder session details were found; clean up stale TEST_DATABASE_URL sessions and rerun"
	}

	parts := make([]string, 0, len(e.Details))
	for _, detail := range e.Details {
		role := "waiter"
		if detail.Granted {
			role = "holder"
		}
		parts = append(parts, fmt.Sprintf(
			"%s pid=%d db=%s user=%s app=%s state=%s wait=%s/%s xact_age=%s query_age=%s query=%q",
			role,
			detail.PID,
			fallback(detail.Database),
			fallback(detail.User),
			fallback(detail.Application),
			fallback(detail.State),
			fallback(detail.WaitEventType),
			fallback(detail.WaitEvent),
			fallback(detail.TransactionAge),
			fallback(detail.QueryAge),
			fallback(detail.Query),
		))
	}

	return message + "; clean up stale TEST_DATABASE_URL sessions and rerun; advisory-lock sessions: " + strings.Join(parts, "; ")
}

func OpenFromEnv(t *testing.T) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), DefaultSetupTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	return db
}

func MustAcquireAdvisoryLock(t *testing.T, ctx context.Context, db *sql.DB, lockKey int64) *sql.Conn {
	t.Helper()

	lockConn, err := AcquireAdvisoryLock(ctx, db, lockKey)
	if err != nil {
		t.Fatalf("acquire test database lock: %v", err)
	}

	t.Cleanup(func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = lockConn.ExecContext(releaseCtx, `SELECT pg_advisory_unlock($1)`, lockKey)
		_ = lockConn.Close()
	})

	return lockConn
}

func AcquireAdvisoryLock(ctx context.Context, db *sql.DB, lockKey int64) (*sql.Conn, error) {
	lockConn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("open test database lock connection: %w", err)
	}

	start := time.Now()
	for {
		var locked bool
		if err := lockConn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&locked); err != nil {
			_ = lockConn.Close()
			return nil, fmt.Errorf("try advisory lock: %w", err)
		}
		if locked {
			return lockConn, nil
		}

		timer := time.NewTimer(lockRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			details, detailErr := loadAdvisoryLockDetails(db, lockKey)
			_ = lockConn.Close()
			return nil, &AdvisoryLockTimeoutError{
				LockKey: lockKey,
				Waited:  time.Since(start),
				Details: details,
				Cause:   firstErr(ctx.Err(), detailErr),
			}
		case <-timer.C:
		}
	}
}

func loadAdvisoryLockDetails(db *sql.DB, lockKey int64) ([]AdvisoryLockSession, error) {
	diagnosticCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	classID, objectID := advisoryLockIDs(lockKey)

	const query = `
SELECT
	l.granted,
	COALESCE(a.pid, 0),
	COALESCE(a.datname, ''),
	COALESCE(a.usename, ''),
	COALESCE(a.application_name, ''),
	COALESCE(a.state, ''),
	COALESCE(a.wait_event_type, ''),
	COALESCE(a.wait_event, ''),
	COALESCE(age(clock_timestamp(), a.xact_start)::text, ''),
	COALESCE(age(clock_timestamp(), a.query_start)::text, ''),
	LEFT(REGEXP_REPLACE(COALESCE(a.query, ''), '\s+', ' ', 'g'), 160)
FROM pg_locks l
LEFT JOIN pg_stat_activity a ON a.pid = l.pid
WHERE l.locktype = 'advisory'
  AND l.classid = $1
  AND l.objid = $2
  AND l.objsubid = 1
ORDER BY l.granted DESC, a.query_start NULLS LAST, a.pid;`

	rows, err := db.QueryContext(diagnosticCtx, query, classID, objectID)
	if err != nil {
		return nil, fmt.Errorf("load advisory lock diagnostics: %w", err)
	}
	defer rows.Close()

	var details []AdvisoryLockSession
	for rows.Next() {
		var detail AdvisoryLockSession
		if err := rows.Scan(
			&detail.Granted,
			&detail.PID,
			&detail.Database,
			&detail.User,
			&detail.Application,
			&detail.State,
			&detail.WaitEventType,
			&detail.WaitEvent,
			&detail.TransactionAge,
			&detail.QueryAge,
			&detail.Query,
		); err != nil {
			return nil, fmt.Errorf("scan advisory lock diagnostics: %w", err)
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate advisory lock diagnostics: %w", err)
	}

	return details, nil
}

func advisoryLockIDs(lockKey int64) (int32, int32) {
	unsigned := uint64(lockKey)
	return int32(unsigned >> 32), int32(unsigned)
}

func fallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func firstErr(primary error, secondary error) error {
	if primary != nil {
		return primary
	}
	return secondary
}
