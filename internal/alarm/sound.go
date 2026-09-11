package alarm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Bell writes the terminal bell. Whether that is audible at all is up to the
// terminal's own settings, which is why it is the fallback rather than the
// first choice.
func Bell(w io.Writer) {
	fmt.Fprint(w, "\a")
}

// Play sounds one strike of the bowl through the system's audio player,
// falling back to Bell where there is no player to be found or it fails.
//
// The setup happens once. The answer cannot change while an alarm rings, and
// rendering the sound and walking PATH again for every strike would be waste.
var Play = func() func(io.Writer) {
	argv := sync.OnceValue(findPlayer)

	return func(w io.Writer) {
		cmd := argv()
		if cmd == nil {
			Bell(w)
			return
		}
		// A strike rings for longer than the gap to the next one, and the
		// player only exits when the sound ends. Waiting for it would let the
		// player set the pace, so start it and leave it, and let the ticker
		// do the timing. Wait is still called, in a goroutine, or every
		// strike would leave a zombie behind.
		play := exec.Command(cmd[0], cmd[1:]...)
		if err := play.Start(); err != nil {
			Bell(w)
			return
		}
		go play.Wait()
	}
}()

// players are the ways to sound a strike, best first. Each is a command line
// missing its last argument, the sound file, which findPlayer supplies.
var players = [][]string{
	{"afplay"}, // macOS
	{"paplay"},
	{"aplay", "-q"},
}

// findPlayer writes the bowl to a file and returns the first command line
// whose player is on PATH, or nil when the machine has no player or nowhere
// to put the file.
func findPlayer() []string {
	for _, cmd := range players {
		if _, err := exec.LookPath(cmd[0]); err != nil {
			continue
		}
		path, err := writeBowl()
		if err != nil {
			return nil
		}
		return append(cmd, path)
	}
	return nil
}

// writeBowl puts the rendered strike in the temp directory under a fixed
// name, so repeated runs reuse one file rather than leaving one each.
//
// Two timers finishing at once would both write it. Writing beside the final
// name and renaming into place keeps a player in the other process from
// reading a half-written file.
func writeBowl() (string, error) {
	final := filepath.Join(os.TempDir(), "timer-bowl.wav")

	tmp, err := os.CreateTemp(filepath.Dir(final), "timer-bowl-*.wav")
	if err != nil {
		return "", err
	}
	if _, err := tmp.Write(bowl()); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	if err := os.Rename(tmp.Name(), final); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return final, nil
}
