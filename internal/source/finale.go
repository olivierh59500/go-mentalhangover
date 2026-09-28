package source

import (
	"encoding/binary"
	"image"
)

type FinaleData struct {
	Logo      *image.NRGBA
	Columns   []byte
	Colors    []uint16
	Balls     []*image.NRGBA
	Particles [][4]byte
	Bounce    []int8
}

// FinalePart keeps the original four-plane logo, perspective column mask,
// ninety-color ramp, six sphere materials and their size rows from cylinder 49.
func FinalePart(data []byte) (FinaleData, error) {
	var result FinaleData
	s := segment{data, 0x9000}
	bank, err := s.read(0x98ac, 30)
	if err != nil {
		return result, err
	}
	palette := make([]uint16, 16)
	for i := 1; i < 16; i++ {
		palette[i] = binary.BigEndian.Uint16(bank[(i-1)*2:])
	}
	bank, err = s.read(0x9b0e, 160*87)
	if err != nil {
		return result, err
	}
	result.Logo, err = DecodePlanar(bank, Planar{Width: 320, Height: 87, RowStride: 160, PlaneOffsets: []int{0, 40, 80, 120}, Palette: palette, TransparentZero: true})
	if err != nil {
		return result, err
	}
	// The original prepares its fifth plane by inserting a cleared longword
	// before each of 112 rows. Source and destination overlap, copying backwards.
	memory := append([]byte(nil), data...)
	dst, src := 0xebac-0x9000, 0xe9ee-0x9000
	for row := 0; row < 112; row++ {
		for word := 0; word < 11; word++ {
			dst -= 4
			src -= 4
			copy(memory[dst:dst+4], memory[src:src+4])
		}
		dst -= 4
		clear(memory[dst : dst+4])
	}
	result.Columns = append([]byte(nil), memory[0xd170-0x9000:0xd170-0x9000+48*185]...)
	bank, err = s.read(0x98ca, 180)
	if err != nil {
		return result, err
	}
	for i := 0; i < 90; i++ {
		result.Colors = append(result.Colors, binary.BigEndian.Uint16(bank[i*2:]))
	}
	ballPalette := []uint16{0, 0xfff, 0xfea, 0xfc6, 0xda4, 0xb72, 0x840, 0xddd, 0xbbb, 0x999, 0x777, 0x555, 0x333, 0x111, 0, 0}
	// The active ball colors belong to the lower copper section, independently
	// of the logo palette. All sixteen entries are retained from source registers.
	for address := uint32(0x9668); address < 0x98ac; address += 4 {
		reg, err := s.word(address)
		if err != nil {
			return result, err
		}
		if reg >= 0x182 && reg <= 0x19e && reg%2 == 0 {
			word, err := s.word(address + 2)
			if err != nil {
				return result, err
			}
			ballPalette[(reg-0x180)/2] = word
		}
	}
	maskBank, err := s.read(0x2550e, 120*32)
	if err != nil {
		return result, err
	}
	mask, err := DecodePlanar(maskBank, Planar{Width: 960, Height: 32, RowStride: 120, PlaneOffsets: []int{0}, Palette: []uint16{0, 0xfff}, TransparentZero: true})
	if err != nil {
		return result, err
	}
	for material := 0; material < 6; material++ {
		bank, err := s.read(0xed0e+uint32(material*0x3c00), 120*128)
		if err != nil {
			return result, err
		}
		pixels, err := DecodePlanar(bank, Planar{Width: 960, Height: 32, RowStride: 480, PlaneOffsets: []int{0, 120, 240, 360}, Palette: ballPalette})
		if err != nil {
			return result, err
		}
		for y := 0; y < 32; y++ {
			for x := 0; x < 960; x++ {
				pixels.Pix[pixels.PixOffset(x, y)+3] = mask.Pix[mask.PixOffset(x, y)+3]
			}
		}
		result.Balls = append(result.Balls, pixels)
	}
	bank, err = s.read(0xebae, 24*4)
	if err != nil {
		return result, err
	}
	for i := 0; i < 24; i++ {
		result.Particles = append(result.Particles, [4]byte{bank[i*4], bank[i*4+1], bank[i*4+2], bank[i*4+3]})
	}
	bank, err = s.read(0xec0e, 256)
	if err != nil {
		return result, err
	}
	for _, v := range bank {
		result.Bounce = append(result.Bounce, int8(v))
	}
	return result, nil
}
