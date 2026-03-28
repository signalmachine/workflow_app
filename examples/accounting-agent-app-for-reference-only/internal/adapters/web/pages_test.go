package web

import (
	"testing"
	"time"
)

func TestFiscalYearBadge_DefaultAprilStart(t *testing.T) {
	t.Setenv("FISCAL_YEAR_START_MONTH", "")

	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{
			name: "before boundary stays previous FY",
			now:  time.Date(2026, time.March, 31, 10, 0, 0, 0, time.UTC),
			want: "FY 2025-26",
		},
		{
			name: "boundary date rolls to new FY",
			now:  time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
			want: "FY 2026-27",
		},
		{
			name: "end of calendar year remains same FY start",
			now:  time.Date(2026, time.December, 31, 23, 59, 0, 0, time.UTC),
			want: "FY 2026-27",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fiscalYearBadge(tt.now)
			if got != tt.want {
				t.Fatalf("fiscalYearBadge(%s) = %q, want %q", tt.now.Format("2006-01-02"), got, tt.want)
			}
		})
	}
}

func TestFiscalYearBadge_RespectsCustomStartMonth(t *testing.T) {
	t.Setenv("FISCAL_YEAR_START_MONTH", "1")

	got := fiscalYearBadge(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	want := "FY 2026-27"
	if got != want {
		t.Fatalf("fiscalYearBadge(custom start) = %q, want %q", got, want)
	}
}

func TestFiscalYearStartMonth_InvalidEnvFallsBack(t *testing.T) {
	t.Setenv("FISCAL_YEAR_START_MONTH", "not-a-number")
	if got := fiscalYearStartMonth(); got != 4 {
		t.Fatalf("fiscalYearStartMonth(invalid) = %d, want 4", got)
	}

	t.Setenv("FISCAL_YEAR_START_MONTH", "13")
	if got := fiscalYearStartMonth(); got != 4 {
		t.Fatalf("fiscalYearStartMonth(out of range) = %d, want 4", got)
	}
}
