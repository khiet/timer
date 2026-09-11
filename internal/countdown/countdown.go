// Package countdown draws a live countdown on one line of a terminal.
package countdown

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// barWidth is the number of cells in the progress bar, chosen to leave room
// for the clock in an 80 column terminal.
const barWidth = 24

// Run redraws the remaining time once a second and returns when the deadline
// passes. If the caller cancels ctx first it returns ctx.Err() and leaves the
// last frame on screen.
//
// The cursor is always left on a fresh line so whatever prints next is not
// overwritten.
func Run(ctx context.Context, w io.Writer, fire time.Time) error {
	total := time.Until(fire)
	defer fmt.Fprintln(w)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		remaining := time.Until(fire)
		if remaining <= 0 {
			draw(w, Line(0, total))
			return nil
		}
		draw(w, Line(remaining, total))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// draw overwrites the current terminal line. \r returns the cursor to column
// one and \033[K clears to end of line, so a frame that is shorter than the
// last one cannot leave stale characters behind.
func draw(w io.Writer, frame string) {
	fmt.Fprintf(w, "\r\033[K%s", frame)
}

// Line formats one frame. It is a pure function of the two durations so the
// display rules can be tested without a clock or a terminal.
func Line(remaining, total time.Duration) string {
	return strings.TrimRight(Clock(remaining)+"  "+bar(remaining, total), " ")
}

// Clock renders d as mm:ss, growing to h:mm:ss once there is an hour to show.
// Part seconds round up, so the display only reads 00:00 at the very end.
func Clock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int((d + time.Second - 1) / time.Second)

	if h := secs / 3600; h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, secs%3600/60, secs%60)
	}
	return fmt.Sprintf("%02d:%02d", secs/60, secs%60)
}

// bar shows elapsed progress. A total of zero or less means the timer was
// already over when it started, so there is no progress worth drawing.
func bar(remaining, total time.Duration) string {
	if total <= 0 {
		return ""
	}
	filled := int(float64(barWidth) * float64(total-remaining) / float64(total))
	filled = min(max(filled, 0), barWidth)
	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled) + "]"
}
