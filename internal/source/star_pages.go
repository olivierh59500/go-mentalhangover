package source

import (
	"encoding/binary"
	"image"
)

type StarPages struct {
	Points     []Point3
	Sines      []int16
	Font       *image.NRGBA
	Advances   []int
	Cards      []TextCard
	StarColors [3]uint16
}

// StarPageData decodes the cylinder-39 steering field and its two text pages.
// The historical extraction filename is not the visual production-unit name.
func StarPageData(data []byte) (StarPages, error) {
	var result StarPages
	s := segment{data, 0x9000}
	bank, err := s.read(0x9a38, 260*6)
	if err != nil {
		return result, err
	}
	for i := 0; i < 260; i++ {
		result.Points = append(result.Points, Point3{int16(binary.BigEndian.Uint16(bank[i*6:])),
			int16(binary.BigEndian.Uint16(bank[i*6+2:])), int16(binary.BigEndian.Uint16(bank[i*6+4:]))})
	}
	bank, err = s.read(0xa398, 720)
	if err != nil {
		return result, err
	}
	for i := 0; i < 360; i++ {
		result.Sines = append(result.Sines, int16(binary.BigEndian.Uint16(bank[i*2:])))
	}
	bank, err = s.read(0x96c6, 118)
	if err != nil {
		return result, err
	}
	for i := 0; i < 59; i++ {
		result.Advances = append(result.Advances, int(binary.BigEndian.Uint16(bank[i*2:]))+1)
	}
	bank, err = s.read(0xa668, 120*28)
	if err != nil {
		return result, err
	}
	result.Font, err = DecodePlanar(bank, Planar{Width: 960, Height: 14, RowStride: 240,
		PlaneOffsets: []int{0, 120}, Palette: []uint16{0, 0xfff, 0x48f, 0x14a}, TransparentZero: true})
	if err != nil {
		return result, err
	}
	for _, d := range []struct {
		name    string
		address uint32
	}{{"special-greetings", 0x973c}, {"members", 0x98ae}} {
		card, err := ReadTextCard(data, 0x9000, d.address, d.name)
		if err != nil {
			return result, err
		}
		result.Cards = append(result.Cards, card)
	}
	for i := range result.StarColors {
		word, err := s.word(0x9a2c + uint32(i*2))
		if err != nil {
			return result, err
		}
		result.StarColors[i] = word
	}
	return result, nil
}
