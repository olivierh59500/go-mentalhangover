package demo

import (
	"fmt"

	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// solidMatrix extends the credit matrix with the third point coordinate. The
// same signed intermediate truncations are used by the original filled BOBs.
func solidMatrix(angles [3]int16, sines []int16) [9]int16 {
	state := [6]int16{angles[0], angles[1], angles[2]}
	first := vectorMatrix(state, sines)
	var wave [6]int16
	for i := 0; i < 3; i++ {
		offset := int(uint16(angles[i]) & 0x0ffe)
		wave[i*2] = sines[offset/2]
		wave[i*2+1] = sines[((offset+0x400)&0x0ffe)/2]
	}
	truncate := func(value int32) int16 { return int16(int32(int16(value>>16)) >> 6) }
	// The solid routine multiplies Z sine by X sine before the high-word
	// truncation. The two-coordinate credit routine uses Y sine first instead.
	// Algebraic commutativity does not survive that intermediate truncation.
	first[4] = truncate(int32(wave[1])*int32(wave[5]) +
		int32(int16((int32(wave[4])*int32(wave[0])<<1)>>16))*int32(wave[2]))
	return [9]int16{first[0], first[1], first[2], first[3], first[4], first[5],
		-(wave[2] >> 7), -truncate(int32(wave[3]) * int32(wave[4])),
		truncate(int32(wave[3]) * int32(wave[5]))}
}

func projectBOB(points []source.Point3, matrix [9]int16, depth int16, output []source.Point2) error {
	if len(points) != len(output) {
		return fmt.Errorf("BOB projection buffer differs from source geometry")
	}
	for i, point := range points {
		x, y, z := int32(point.X), int32(point.Y), int32(point.Z)
		px := x*int32(matrix[0]) + y*int32(matrix[3]) + z*int32(matrix[6])
		py := x*int32(matrix[1]) + y*int32(matrix[4]) + z*int32(matrix[7])
		pz := x*int32(matrix[2]) + y*int32(matrix[5]) + z*int32(matrix[8])
		divisor := int16(uint16(int16(pz>>9)) + uint16(depth))
		if divisor == 0 {
			return fmt.Errorf("original BOB projection divides by zero")
		}
		divide := func(value int32) int16 {
			quotient := value / int32(divisor)
			if quotient < -32768 || quotient > 32767 {
				return int16(value)
			}
			return int16(quotient)
		}
		output[i] = source.Point2{X: int16(uint16(divide(px)) + 15), Y: int16(uint16(divide(py)) + 14)}
	}
	return nil
}
