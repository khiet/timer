package deadline_test

import (
	"testing"
	"time"

	"github.com/khiet/timer/internal/deadline"
)

// now is the fixed clock every case below is resolved against.
var now = time.Date(2027, time.March, 14, 8, 0, 0, 0, time.Local)

// at builds a time on the same day as now.
func at(hour, min int) time.Time {
	return time.Date(2027, time.March, 14, hour, min, 0, 0, time.Local)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want time.Time
	}{
		{"hours and minutes", "2h15m", now.Add(135 * time.Minute)},
		{"minutes", "45m", now.Add(45 * time.Minute)},
		{"minutes and seconds", "8m20s", now.Add(8*time.Minute + 20*time.Second)},
		{"seconds alone", "90s", now.Add(90 * time.Second)},
		{"bare number means seconds", "10", now.Add(10 * time.Second)},
		{"bare fraction means seconds", "1.5", now.Add(1500 * time.Millisecond)},
		{"absolute date", "2027-03-14T09:00", at(9, 0)},
		{"absolute date with seconds", "2027-03-14T09:00:30", at(9, 0).Add(30 * time.Second)},
		{"time of day later today", "T09:00", at(9, 0)},
		{"time of day already passed rolls over", "T07:30", at(7, 30).AddDate(0, 0, 1)},
		{"lowercase t is accepted", "t09:00", at(9, 0)},
		{"surrounding space is ignored", " 45m ", now.Add(45 * time.Minute)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deadline.Parse(tt.arg, now)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.arg, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("Parse(%q) = %s, want %s", tt.arg, got, tt.want)
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{"empty", ""},
		{"nonsense", "banana"},
		{"bare zero", "0"},
		{"number with a stray unit", "45x"},
		{"negative duration", "-5m"},
		{"zero duration", "0s"},
		{"date in the past", "2020-01-01T00:00"},
		{"impossible time of day", "T25:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := deadline.Parse(tt.arg, now); err == nil {
				t.Errorf("Parse(%q) = %s, want an error", tt.arg, got)
			}
		})
	}
}

// The countdown relies on this, so assert it once across every valid form.
func TestParseAlwaysReturnsAFutureTime(t *testing.T) {
	for _, arg := range []string{"2h15m", "45m", "8m20s", "2027-03-14T09:00", "T09:00", "T07:30"} {
		fire, err := deadline.Parse(arg, now)
		if err != nil {
			t.Fatalf("Parse(%q): %v", arg, err)
		}
		if !fire.After(now) {
			t.Errorf("Parse(%q) = %s, which is not after now (%s)", arg, fire, now)
		}
	}
}
