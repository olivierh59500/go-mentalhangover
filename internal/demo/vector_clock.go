package demo

import (
	"fmt"

	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// vectorClock mirrors word-sized self-modifying cue operands at 0x932a.
// New deltas apply on the cue-boundary frame, before rotation and projection.
type vectorClock struct {
	model       source.VectorModel
	state       [6]int16
	cue, within int
	started     bool
	done        bool
}

func newVectorClock(model source.VectorModel) *vectorClock {
	return &vectorClock{model: model, state: model.Initial}
}

func (clock *vectorClock) Step() bool {
	if clock.done {
		return false
	}
	if !clock.started {
		clock.started = true
	} else {
		clock.within++
		if clock.within >= clock.model.Cues[clock.cue].Frames {
			clock.cue++
			clock.within = 0
		}
	}
	if clock.cue >= len(clock.model.Cues) {
		clock.done = true
		return false
	}
	for i, delta := range clock.model.Cues[clock.cue].Delta {
		clock.state[i] = int16(uint16(clock.state[i]) + uint16(delta))
	}
	return true
}

// vectorMatrix retains signed high-word truncation, including intermediate
// word products. Replacing this with floating Euler angles changes source pixels.
func vectorMatrix(state [6]int16, sines []int16) [6]int16 {
	var wave [6]int16
	for i := 0; i < 3; i++ {
		offset := int(uint16(state[i]) & 0x0ffe)
		wave[i*2] = sines[offset/2]
		wave[i*2+1] = sines[((offset+0x400)&0x0ffe)/2]
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(value int32) int16 { return int16(value >> 16) }
	truncate := func(value int32) int16 { return int16(int32(high(value)) >> 6) }
	return [6]int16{
		truncate(mul(cx, cy)),
		truncate(mul(sx, cz) - mul(high(mul(sz, sy)<<1), cx)),
		truncate(mul(sx, sz) + mul(high(mul(cz, sy)<<1), cx)),
		-truncate(mul(sx, cy)),
		truncate(mul(cx, cz) + mul(high(mul(sz, sy)<<1), sx)),
		truncate(mul(cx, sz) - mul(high(mul(cz, sy)<<1), sx)),
	}
}

// projectVector emulates DIVS word results and the production's (175,136) origin.
func projectVector(points []source.Point2, matrix [6]int16, state [6]int16, output []source.Point2) error {
	if len(output) != len(points) {
		return fmt.Errorf("vector projection buffer differs from point bank")
	}
	divide := func(value int32, denominator int16) int16 {
		quotient := value / int32(denominator)
		if quotient < -32768 || quotient > 32767 {
			// DIVS overflow leaves the destination unchanged on the 68000.
			return int16(value)
		}
		return int16(quotient)
	}
	for i, point := range points {
		x, y := int32(point.X), int32(point.Y)
		px := x*int32(matrix[0]) + y*int32(matrix[3])
		py := x*int32(matrix[1]) + y*int32(matrix[4])
		pz := x*int32(matrix[2]) + y*int32(matrix[5])
		depth := int16(pz >> 9)
		depth = int16(uint16(depth) + 512 + uint16(state[5]>>1))
		if depth == 0 {
			return fmt.Errorf("original vector projection divides by zero")
		}
		px += int32(state[3]) << 8
		py += int32(state[4]) << 8
		output[i] = source.Point2{X: int16(uint16(divide(px, depth)) + 175),
			Y: int16(uint16(divide(py, depth)) + 136)}
	}
	return nil
}
