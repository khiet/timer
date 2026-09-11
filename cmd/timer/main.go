// Command timer counts down to a deadline and then strikes a singing bowl.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khiet/timer/internal/alarm"
	"github.com/khiet/timer/internal/countdown"
	"github.com/khiet/timer/internal/deadline"
)

const usage = `timer counts down and then rings until you press Enter.

usage: timer <duration|time>

  timer 2h15m             two hours and a quarter from now
  timer 45m               three quarters of an hour from now
  timer 8m20s             eight minutes and twenty seconds from now
  timer 10                ten seconds from now; a bare number means seconds
  timer T07:30            the next 07:30, today or tomorrow
  timer 2027-03-14T09:00  that exact local date and time

Press Ctrl-C to cancel.
`

// main stays three lines long. Everything it would otherwise do lives in run,
// which takes the process globals as arguments, so the whole program can be
// exercised from a test with nothing more than a buffer and a string.
func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "timer:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, stdin io.Reader) error {
	flags := flag.NewFlagSet("timer", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stderr, usage) }

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil // asking for help is not a failure
		}
		return err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return errors.New("expected one duration or time")
	}

	fire, err := deadline.Parse(flags.Arg(0), time.Now())
	if err != nil {
		return err
	}

	// Ctrl-C cancels the countdown and silences the alarm. Restoring the
	// default handler on the way out means a second Ctrl-C still kills a
	// process that has somehow wedged.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(stdout, "Timer set for %s\n", fire.Format("Mon 15:04:05"))

	if err := countdown.Run(ctx, stdout, fire); err != nil {
		return quiet(err)
	}

	fmt.Fprintln(stdout, "Time is up. Press Enter to dismiss.")
	return quiet(alarm.Default().Ring(ctx, stdout, stdin))
}

// quiet swallows the error a Ctrl-C produces. The user asked for the program
// to stop, so stopping is a success; anything else is passed through.
func quiet(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
