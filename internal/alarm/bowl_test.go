package alarm

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// The bowl is code, so these are the things a listener would otherwise have
// to check by ear: it is a well-formed file, it rings for about five seconds,
// it neither clips nor whispers, and it ends in silence.

func TestBowlIsAPCMWav(t *testing.T) {
	wav := bowl()

	if !bytes.HasPrefix(wav, []byte("RIFF")) || string(wav[8:16]) != "WAVEfmt " || string(wav[36:40]) != "data" {
		t.Fatal("missing RIFF/WAVE/fmt/data markers")
	}
	le := binary.LittleEndian
	if got := le.Uint32(wav[4:]); got != uint32(len(wav)-8) {
		t.Errorf("RIFF size = %d, want %d", got, len(wav)-8)
	}
	if got := le.Uint32(wav[40:]); got != uint32(len(wav)-44) {
		t.Errorf("data size = %d, want %d", got, len(wav)-44)
	}
	if le.Uint16(wav[20:]) != 1 || le.Uint16(wav[22:]) != 1 || le.Uint16(wav[34:]) != 16 {
		t.Error("want PCM, mono, 16-bit")
	}
	if got := le.Uint32(wav[24:]); got != sampleRate {
		t.Errorf("sample rate = %d, want %d", got, sampleRate)
	}
}

func TestBowlRingsThenFallsSilent(t *testing.T) {
	wav := bowl()
	samples := wav[44:]
	n := len(samples) / 2

	seconds := float64(n) / sampleRate
	if seconds < 5 || seconds > 6 {
		t.Errorf("rings for %.1fs, want about five", seconds)
	}

	peak := func(from, to int) int {
		best := 0
		for i := from; i < to; i++ {
			v := int(int16(binary.LittleEndian.Uint16(samples[2*i:])))
			if v < 0 {
				v = -v
			}
			best = max(best, v)
		}
		return best
	}

	// Full scale is 32767: loud enough to hear, with headroom against
	// clipping in a player that adds its own gain.
	if p := peak(0, n); p < 20000 || p > 30000 {
		t.Errorf("peak = %d, want between 20000 and 30000", p)
	}
	// The strike is at the start, and the last tenth of a second is the tail.
	if first, last := peak(0, sampleRate/10), peak(n-sampleRate/10, n); last*20 > first {
		t.Errorf("last tenth peaks at %d against %d at the strike; the ring should have died away", last, first)
	}
}

func TestBowlIsDeterministic(t *testing.T) {
	if !bytes.Equal(bowl(), bowl()) {
		t.Error("two renders differ")
	}
}
