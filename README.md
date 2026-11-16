# timer

A countdown timer for the terminal. It redraws the time remaining on a single
line, then strikes a singing bowl three times until you acknowledge it. Press
Enter to dismiss the alarm, or Ctrl-C to cancel at any point. The bowl plays
through `afplay`, `paplay` or `aplay`; where none is available it falls back to
the terminal bell.

```
$ timer 45m
Timer set for Fri 18:21:04
31:17  [==========              ]
```

## Install

```
go install github.com/khiet/timer/cmd/timer@latest
```

From a clone, without installing:

```
go run ./cmd/timer 45m
go build -o timer ./cmd/timer
```

## Usage

`timer` takes exactly one argument, how long to count down for:

| Example | Means |
| --- | --- |
| `timer 1h30m` | an hour and a half from now |
| `timer 10s` | ten seconds from now |
| `timer 10` | the same; a bare number means seconds |

Durations are whatever [`time.ParseDuration`][pd] accepts, so `h`, `m` and `s`
combine in any order and finer units work too. There is no unit for days;
`timer 36h` is the way to ask for a day and a half.

[pd]: https://pkg.go.dev/time#ParseDuration

## Layout

```
cmd/timer/main.go       argument handling and wiring
internal/deadline/      what "45m" and "T07:30" mean, as pure functions
internal/countdown/     the live display
internal/alarm/         the strike pattern, its sound and its acknowledgement
```

Four conventions are doing the work here, and they are the interesting part of
the code:

**`main` is three lines.** It calls `run(args, stdout, stderr, stdin) error`
and turns an error into an exit code. Because `run` is handed the process
globals rather than reaching for them, the whole command is testable with a
buffer and a string. See `cmd/timer/main_test.go`.

**`cmd/` and `internal/`.** `cmd/timer` is the binary; `internal/` is code the
compiler will refuse to let anyone outside this module import. That refusal is
what keeps a package private until you decide otherwise. (The widely linked
`golang-standards/project-layout` is not an official standard, and its `pkg/`
directory in particular buys nothing.)

**Packages are named for what they are about, not what layer they are.** There
is no `utils`, `helpers` or `models`. `deadline.Parse` reads well at the call
site precisely because the package name is half the sentence.

**The pure part is separated from the I/O.** `deadline.Parse` takes `now` as an
argument instead of calling `time.Now`, and `countdown.Clock` formats a
duration without touching a terminal. Both are then testable with a table and
no mocks at all. The two places that genuinely need a seam take one: reading
the acknowledgement takes an `io.Reader` and gets `strings.NewReader("\n")` in
its test, and `alarm.Alarm.Tick` is a function the test replaces with a counter
so `go test ./...` stays silent.

## Tests

```
go test ./...
```

