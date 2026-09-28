package demo

import (
	"fmt"

	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func patternedMatrix(angles [3]int16, sines []int16) [9]int16 {
	var wave [6]int16
	for i, angle := range angles {
		offset := int(angle) &^ 1
		wave[i*2] = sines[offset/2]
		wave[i*2+1] = sines[((offset+180)%720)/2]
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(n int32) int16 { return int16(n >> 16) }
	truncate := func(n int32) int16 { return int16(int32(high(n)) >> 7) }
	return [9]int16{
		truncate(mul(cx, cy)), truncate(mul(sx, cz) - mul(high(mul(sz, sy)<<1), cx)),
		truncate(mul(sx, sz) + mul(high(mul(cz, sy)<<1), cx)),
		-truncate(mul(sx, cy)), truncate(mul(cx, cz) + mul(high(mul(sz, sx)<<1), sy)),
		truncate(mul(cx, sz) - mul(high(mul(cz, sy)<<1), sx)),
		-(sy >> 8), -truncate(mul(cy, sz)), truncate(mul(cy, cz)),
	}
}

func projectPatterned(points []source.Point3, matrix [9]int16, state [6]int16, output []source.Point2) error {
	if len(points) != len(output) {
		return fmt.Errorf("patterned projection buffer differs from geometry")
	}
	for i, p := range points {
		x, y, z := int32(p.X), int32(p.Y), int32(p.Z)
		px := x*int32(matrix[0]) + y*int32(matrix[3]) + z*int32(matrix[6])
		py := x*int32(matrix[1]) + y*int32(matrix[4]) + z*int32(matrix[7])
		pz := x*int32(matrix[2]) + y*int32(matrix[5]) + z*int32(matrix[8])
		depth := int16(uint16(int16(pz>>9)) + uint16(state[5]) + 512)
		if depth == 0 {
			return fmt.Errorf("original patterned solid divides by zero")
		}
		divide := func(n int32) int16 {
			q := n / int32(depth)
			if q < -32768 || q > 32767 {
				return int16(n)
			}
			return int16(q)
		}
		px += int32(state[3]) << 7
		py += int32(state[4]) << 7
		output[i] = source.Point2{X: int16(uint16(divide(px)) + 256), Y: int16(uint16(divide(py)) + 100)}
	}
	return nil
}
