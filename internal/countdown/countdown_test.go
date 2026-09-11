package countdown_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/khiet/timer/internal/countdown"
)

func TestClock(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"zero", 0, "00:00"},
		{"seconds", 30 * time.Second, "00:30"},
		{"minutes", 45 * time.Minute, "45:00"},
		{"minutes and seconds", 8*time.Minute + 20*time.Second, "08:20"},
		{"an hour grows the field", 135 * time.Minute, "2:15:00"},
		{"part seconds round up", 1500 * time.Millisecond, "00:02"},
		{"negative clamps to zero", -5 * time.Second, "00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countdown.Clock(tt.in); got != tt.want {
				t.Errorf("Clock(%s) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestLineFillsTheBarAsTimePasses(t *testing.T) {
	total := 10 * time.Minute

	start := countdown.Line(total, total)
	half := countdown.Line(total/2, total)
	end := countdown.Line(0, total)

	if strings.Count(start, "=") != 0 {
		t.Errorf("at the start the bar should be empty, got %q", start)
	}
	if got, want := strings.Count(half, "="), strings.Count(end, "=")/2; got != want {
		t.Errorf("halfway the bar should be half full, got %q", half)
	}
	if strings.Contains(end, " ]") {
		t.Errorf("at the end the bar should be full, got %q", end)
	}
}

func TestRunReturnsWhenTheDeadlineHasPassed(t *testing.T) {
	var buf bytes.Buffer

	if err := countdown.Run(context.Background(), &buf, time.Now()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "00:00") {
		t.Errorf("final frame should read 00:00, got %q", got)
	}
}

func TestRunStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := countdown.Run(ctx, &bytes.Buffer{}, time.Now().Add(time.Hour))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run returned %v, want context.Canceled", err)
	}
}
