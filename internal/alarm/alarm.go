// Package alarm makes the noise at the end of a timer.
package alarm

import (
	"bufio"
	"context"
	"io"
	"time"
)

// An Alarm is a strike pattern: when the time is up it strikes a singing
// bowl a fixed number of times, and then gives up unless someone acknowledges
// it first.
//
// The zero value never strikes; use Default.
type Alarm struct {
	Repeats  int             // how many strikes before giving up
	Interval time.Duration   // gap between strikes
	Tick     func(io.Writer) // sounds one strike; nil means the terminal bell
}

// Default is the pattern the command uses: three strikes, three seconds
// apart. Each strike rings for about five seconds, so they overlap and the
// whole alarm fades out rather than stopping.
func Default() Alarm {
	return Alarm{Repeats: 3, Interval: 3 * time.Second, Tick: Play}
}

// Ring strikes until the pattern is exhausted, the user presses Enter on ack, or
// ctx is cancelled. It returns ctx.Err() only in the last case.
//
// w carries the terminal bell when there is no audio player to sound a strike.
//
// ack is normally os.Stdin. When it is not a terminal (output piped to a file,
// say) it reports end of input immediately; that is not an acknowledgement, so
// the alarm rings out in full.
func (a Alarm) Ring(ctx context.Context, w io.Writer, ack io.Reader) error {
	acked := waitForEnter(ack)

	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	for range a.Repeats {
		a.tick(w)

		select {
		case <-acked:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	return nil
}

// tick sounds one strike. An Alarm built by hand carries no Tick, and the bell
// is the one noise available without asking anything of the machine.
func (a Alarm) tick(w io.Writer) {
	if a.Tick == nil {
		Bell(w)
		return
	}
	a.Tick(w)
}

// waitForEnter reports a single line from r on the returned channel.
//
// A read blocked on a terminal cannot be interrupted, so this goroutine
// outlives Ring whenever nobody types. The channel is buffered for exactly
// that case: the send completes with no receiver and the goroutine ends,
// rather than leaking until the process exits.
func waitForEnter(r io.Reader) <-chan struct{} {
	acked := make(chan struct{}, 1)
	go func() {
		if _, err := bufio.NewReader(r).ReadString('\n'); err == nil {
			acked <- struct{}{}
		}
	}()
	return acked
}
