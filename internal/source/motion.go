package source

import (
	"encoding/binary"
	"fmt"
)

// ReadMotion preserves the six-channel word program shared by credits and
// vector units: initial pose, frame count plus deltas, then a 0x7fff sentinel.
func ReadMotion(data []byte, base, address uint32) ([6]int16, []VectorCue, error) {
	var initial [6]int16
	s := segment{data, base}
	bank, err := s.read(address, 12)
	if err != nil {
		return initial, nil, err
	}
	for i := range initial {
		initial[i] = int16(binary.BigEndian.Uint16(bank[i*2:]))
	}
	address += 12
	var cues []VectorCue
	for i := 0; i < 1024; i++ {
		n, err := s.word(address)
		if err != nil {
			return initial, nil, err
		}
		if n == 0x7fff {
			return initial, cues, nil
		}
		if n < 1 || n > 16384 {
			return initial, nil, fmt.Errorf("source: invalid word motion duration at %x", address)
		}
		bank, err = s.read(address+2, 12)
		if err != nil {
			return initial, nil, err
		}
		cue := VectorCue{Frames: int(n)}
		for j := range cue.Delta {
			cue.Delta[j] = int16(binary.BigEndian.Uint16(bank[j*2:]))
		}
		cues = append(cues, cue)
		address += 14
	}
	return initial, nil, fmt.Errorf("source: word motion sentinel missing")
}
