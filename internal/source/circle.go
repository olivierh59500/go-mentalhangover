package source

import (
	"fmt"
	"image"
)

type PolarGlyphPoint struct{ Angle, Row byte }
type PolarGlyph struct{ Contours [][]PolarGlyphPoint }
type CircleData struct {
	Glyphs           map[byte]PolarGlyph
	Wave, Profile    []int8
	Text             []byte
	Mountain, Raster *image.NRGBA
}

// CirclePart decodes the source outline font, two byte waves, binary control
// text, mountain mask and the 49 paired copper colors from cylinder 44.
func CirclePart(data []byte) (CircleData, error) {
	result := CircleData{Glyphs: make(map[byte]PolarGlyph, 59)}
	s := segment{data, 0x9000}
	for _, address := range []uint32{0x9d12, 0x9e12} {
		bank, err := s.read(address, 256)
		if err != nil {
			return result, err
		}
		wave := make([]int8, 256)
		for i, v := range bank {
			wave[i] = int8(v)
		}
		if address == 0x9d12 {
			result.Wave = wave
		} else {
			result.Profile = wave
		}
	}
	var err error
	result.Glyphs, err = ReadOutlineFont(data, 0x9000, 0x9958, 10)
	if err != nil {
		return result, err
	}
	address := uint32(0x9f12)
	for len(result.Text) < 4096 {
		bank, err := s.read(address, 1)
		if err != nil {
			return result, err
		}
		value := bank[0]
		address++
		result.Text = append(result.Text, value)
		if value == 255 {
			break
		}
		count := 0
		if value == '|' {
			count = 1
		} else if value == '>' {
			count = 5
		} else if value < 32 || value > 90 {
			return result, fmt.Errorf("source: invalid circular text byte")
		}
		if count > 0 {
			bank, err := s.read(address, count)
			if err != nil {
				return result, err
			}
			result.Text = append(result.Text, bank...)
			address += uint32(count)
		}
	}
	if len(result.Text) == 0 || result.Text[len(result.Text)-1] != 255 {
		return result, fmt.Errorf("source: missing circle text terminator")
	}
	bank, err := s.read(0x9fe8, 44*200)
	if err != nil {
		return result, err
	}
	result.Mountain, err = DecodePlanar(bank, Planar{Width: 352, Height: 200, RowStride: 44, PlaneOffsets: []int{0}, Palette: []uint16{0, 0}, TransparentZero: true})
	if err != nil {
		return result, err
	}
	result.Raster = image.NewNRGBA(image.Rect(0, 0, 352, 272))
	for i := 0; i < 49; i++ {
		word, err := s.word(0x9868 + uint32(i*2))
		if err != nil {
			return result, err
		}
		for y := 102 + i*2; y < 104+i*2; y++ {
			for x := 0; x < 352; x++ {
				result.Raster.SetNRGBA(x, y, RGB12(word))
			}
		}
	}
	return result, nil
}
