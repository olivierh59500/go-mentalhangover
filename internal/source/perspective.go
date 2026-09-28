package source

import (
	"encoding/binary"
	"fmt"
)

type PerspectiveData struct {
	Glyphs map[byte]PolarGlyph
	Text   []byte
	Points []Point3
}

func PerspectivePart(data []byte) (PerspectiveData, error) {
	var result PerspectiveData
	var err error
	result.Glyphs, err = ReadOutlineFont(data, 0x9000, 0xae40, 10)
	if err != nil {
		return result, err
	}
	s := segment{data, 0x9000}
	address := uint32(0xb12a)
	for len(result.Text) < 4096 {
		bank, err := s.read(address, 1)
		if err != nil {
			return result, err
		}
		result.Text = append(result.Text, bank[0])
		address++
		if bank[0] == 255 {
			break
		}
		if bank[0] < 32 || bank[0] > 90 {
			return result, fmt.Errorf("source: invalid perspective message byte")
		}
	}
	if len(result.Text) == 0 || result.Text[len(result.Text)-1] != 255 {
		return result, fmt.Errorf("source: missing perspective message end")
	}
	// The source's unrelated point layer shares the display with the outlines.
	bank, err := s.read(0x9aca, 101*6)
	if err != nil {
		return result, err
	}
	for i := 0; i < 101; i++ {
		result.Points = append(result.Points, Point3{int16(binary.BigEndian.Uint16(bank[i*6:])),
			int16(binary.BigEndian.Uint16(bank[i*6+2:])), int16(binary.BigEndian.Uint16(bank[i*6+4:]))})
	}
	return result, nil
}
