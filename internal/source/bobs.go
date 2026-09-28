package source

import (
	"encoding/binary"
	"fmt"
	"image"
)

// BOBToken preserves binary controls embedded in the original scrolling text.
// A lower-case b consumes fourteen bytes, including zeros and non-text values.
type BOBToken struct {
	Kind    byte
	Letter  rune
	Values  [14]byte
	Address uint32
}
type BOBData struct {
	Tokens   []BOBToken
	Text     string
	Advances []int
	Font     *image.NRGBA
	Sines    []int16
	X, Y     []int
	Palette  *image.NRGBA
}

func ReadBOBs(data []byte) (BOBData, error) {
	var result BOBData
	s := segment{data, 0x9000}
	widths, err := s.read(0x9ae2, 118)
	if err != nil {
		return result, err
	}
	for i := 0; i < 59; i++ {
		result.Advances = append(result.Advances, int(binary.BigEndian.Uint16(widths[i*2:]))+1)
	}
	fontData, err := s.read(0x9b58, 120*28)
	if err != nil {
		return result, err
	}
	result.Font, err = DecodePlanar(fontData, Planar{Width: 960, Height: 14, RowStride: 240,
		PlaneOffsets: []int{0, 120}, Palette: []uint16{0, 0xfff, 0x48f, 0x14a}, TransparentZero: true})
	if err != nil {
		return result, err
	}
	bank, err := s.read(0xaed8, 4096)
	if err != nil {
		return result, err
	}
	for i := 0; i < 2048; i++ {
		result.Sines = append(result.Sines, int16(binary.BigEndian.Uint16(bank[i*2:])))
	}
	bank, err = s.read(0xbed8, 2048)
	if err != nil {
		return result, err
	}
	for i := 0; i < 512; i++ {
		// Chip DMA ignores bit zero of word addresses. Keeping that bit would
		// introduce an eight-pixel jump on alternating source table entries.
		result.X = append(result.X, int(bank[i*2]&0xfe)*8+int(bank[i*2+1]))
		result.Y = append(result.Y, int(binary.BigEndian.Uint16(bank[1024+i*2:]))/88)
	}
	result.Palette = image.NewNRGBA(image.Rect(0, 0, 3, 272))
	for y := 0; y < 272; y++ {
		for material := 0; material < 3; material++ {
			word := []uint16{0xfff, 0x48f, 0x14a}[material]
			if y >= 33 {
				word, err = s.word(0xabec + uint32(material*54+min(25, (y-33)/8)*2))
				if err != nil {
					return result, err
				}
			}
			result.Palette.SetNRGBA(material, y, RGB12(word))
		}
	}
	text := make([]rune, 0, 1024)
	address := uint32(0xac86)
	for len(result.Tokens) < 4096 {
		value, err := s.read(address, 1)
		if err != nil {
			return result, err
		}
		address++
		if value[0] == 0 {
			result.Text = string(text)
			return result, nil
		}
		token := BOBToken{Kind: value[0], Address: address - 1}
		switch value[0] {
		case 'b':
			values, err := s.read(address, 14)
			if err != nil {
				return result, err
			}
			copy(token.Values[:], values)
			address += 14
			if token.Values[0] > 64 || token.Values[13] > 1 {
				return result, fmt.Errorf("source: invalid BOB count or model")
			}
		case '>', '|':
			values, err := s.read(address, 1)
			if err != nil {
				return result, err
			}
			if values[0] < '0' || values[0] > '9' {
				return result, fmt.Errorf("source: invalid scrolling BOB control")
			}
			token.Values[0] = values[0] - '0'
			address++
		case 'a', 'd':
		default:
			if value[0] < 32 || value[0] > 90 {
				return result, fmt.Errorf("source: unsupported BOB character %x", value[0])
			}
			token.Kind = 't'
			token.Letter = rune(value[0])
			text = append(text, token.Letter)
		}
		result.Tokens = append(result.Tokens, token)
	}
	return result, fmt.Errorf("source: missing BOB text terminator")
}
