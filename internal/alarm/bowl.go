package alarm

import (
	"encoding/binary"
	"math"
	"math/rand/v2"
)

// The sound is a singing bowl struck once: a low note that rings for a few
// seconds and dies away on its own. It is synthesised here rather than shipped
// as a file so that the repository stays text and the sound can be read.
//
// A bowl's ring is a handful of sine partials that decay at different rates,
// the higher ones faster. The two near-unison pairs (1 and 1.006, 2.72 and
// 2.735) beat slowly against each other, which is the shimmer a real bowl has
// and a plain chord does not.
const (
	sampleRate  = 44100
	fundamental = 196.0 // G3
	ringSeconds = 5.2
)

var partials = []struct {
	ratio, amp, decay float64 // decay is the share of ringSeconds this partial lasts
}{
	{1, 1, 1},
	{1.006, 0.7, 1},
	{2.72, 0.5, 0.75},
	{2.735, 0.3, 0.7},
	{5.1, 0.18, 0.45},
	{8.3, 0.06, 0.25},
}

// bowl renders one strike as a 16-bit mono WAV. It is deterministic: the
// striking noise is seeded, so two runs produce the same bytes.
func bowl() []byte {
	n := int(ringSeconds * sampleRate)
	samples := make([]float64, n)

	const attack = 0.012 // seconds from silence to full ring
	const floor = 0.0004 // the level a partial decays to before it is cut
	for _, p := range partials {
		length := p.decay * ringSeconds
		for i := range samples {
			t := float64(i) / sampleRate
			if t >= length {
				break
			}
			env := p.amp
			if t < attack {
				env *= t / attack
			} else {
				// An exponential ramp from full level at the end of the
				// attack down to floor at the end of the partial.
				env *= math.Pow(floor, (t-attack)/(length-attack))
			}
			samples[i] += env * math.Sin(2*math.Pi*fundamental*p.ratio*t)
		}
	}

	// The strike itself: a brief thump of low-passed noise, gone in 60 ms.
	rng := rand.New(rand.NewPCG(1, 2))
	const thump = 0.06
	const thumpAmp = 0.5
	var lp float64 // one-pole low-pass state, corner near 900 Hz
	alpha := 1 - math.Exp(-2*math.Pi*900/sampleRate)
	for i := 0; float64(i)/sampleRate < thump; i++ {
		t := float64(i) / sampleRate
		lp += alpha * (rng.Float64()*2 - 1 - lp)
		samples[i] += thumpAmp * math.Pow(floor, t/thump) * lp
	}

	// Normalise so the loudest moment sits just below full scale.
	peak := 0.0
	for _, s := range samples {
		peak = math.Max(peak, math.Abs(s))
	}
	scale := 0.85 / peak

	const header = 44
	out := make([]byte, header+2*n)
	le := binary.LittleEndian
	copy(out[0:], "RIFF")
	le.PutUint32(out[4:], uint32(len(out)-8))
	copy(out[8:], "WAVEfmt ")
	le.PutUint32(out[16:], 16)         // fmt chunk size
	le.PutUint16(out[20:], 1)          // PCM
	le.PutUint16(out[22:], 1)          // channels
	le.PutUint32(out[24:], sampleRate) // sample rate
	le.PutUint32(out[28:], sampleRate*2)
	le.PutUint16(out[32:], 2)  // block align
	le.PutUint16(out[34:], 16) // bits per sample
	copy(out[36:], "data")
	le.PutUint32(out[40:], uint32(2*n))
	for i, s := range samples {
		le.PutUint16(out[header+2*i:], uint16(int16(s*scale*math.MaxInt16)))
	}
	return out
}
