package alarm_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/khiet/timer/internal/alarm"
)

// test is the Default pattern with the waiting taken out. Its Tick is nil, so
// it falls back to the bell and the assertions below can count characters
// instead of listening.
var test = alarm.Alarm{Repeats: 3, Interval: time.Millisecond}

func TestRingBeepsTheWholePatternWhenNobodyAcknowledges(t *testing.T) {
	var beeps bytes.Buffer

	if err := test.Ring(context.Background(), &beeps, strings.NewReader("")); err != nil {
		t.Fatalf("Ring returned %v", err)
	}
	if got := strings.Count(beeps.String(), "\a"); got != test.Repeats {
		t.Errorf("rang %d times, want %d", got, test.Repeats)
	}
}

func TestRingStopsOnEnter(t *testing.T) {
	var beeps bytes.Buffer
	slow := alarm.Alarm{Repeats: 100, Interval: time.Millisecond}

	if err := slow.Ring(context.Background(), &beeps, strings.NewReader("\n")); err != nil {
		t.Fatalf("Ring returned %v", err)
	}
	if got := strings.Count(beeps.String(), "\a"); got >= slow.Repeats {
		t.Errorf("rang %d times, want it cut short by the acknowledgement", got)
	}
}

func TestRingUsesTheTickItWasGiven(t *testing.T) {
	ticks := 0
	counted := alarm.Alarm{
		Repeats:  3,
		Interval: time.Millisecond,
		Tick:     func(io.Writer) { ticks++ },
	}

	if err := counted.Ring(context.Background(), &bytes.Buffer{}, strings.NewReader("")); err != nil {
		t.Fatalf("Ring returned %v", err)
	}
	if ticks != counted.Repeats {
		t.Errorf("ticked %d times, want %d", ticks, counted.Repeats)
	}
}

// The point of the pattern: three unhurried strikes, not an oven's nag.
func TestDefaultStrikesThreeTimesThreeSecondsApart(t *testing.T) {
	d := alarm.Default()

	if d.Repeats != 3 {
		t.Errorf("Repeats = %d, want 3", d.Repeats)
	}
	if d.Interval != 3*time.Second {
		t.Errorf("Interval = %s, want 3s", d.Interval)
	}
}

func TestRingStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := test.Ring(ctx, &bytes.Buffer{}, strings.NewReader(""))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Ring returned %v, want context.Canceled", err)
	}
}
