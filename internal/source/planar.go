package source

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// Planar describes the original bitplanes, including interleaved font rows.
type Planar struct {
	Width, Height, RowStride int
	PlaneOffsets             []int
	Palette                  []uint16
	TransparentZero          bool
}

func RGB12(value uint16) color.NRGBA {
	return color.NRGBA{R: uint8(value>>8&15) * 17, G: uint8(value>>4&15) * 17, B: uint8(value&15) * 17, A: 255}
}

// DecodePlanar keeps original bit order and palette indices; no scaling occurs.
func DecodePlanar(data []byte, spec Planar) (*image.NRGBA, error) {
	if spec.Width < 1 || spec.Height < 1 || spec.Width > 8192 || spec.Height > 8192 ||
		spec.RowStride < (spec.Width+7)/8 || len(spec.PlaneOffsets) < 1 || len(spec.PlaneOffsets) > 8 ||
		len(spec.Palette) < 1<<len(spec.PlaneOffsets) {
		return nil, fmt.Errorf("source: invalid planar image")
	}
	for _, offset := range spec.PlaneOffsets {
		if offset < 0 || offset+(spec.Height-1)*spec.RowStride+(spec.Width+7)/8 > len(data) {
			return nil, fmt.Errorf("source: truncated bitplane")
		}
	}
	result := image.NewNRGBA(image.Rect(0, 0, spec.Width, spec.Height))
	for y := 0; y < spec.Height; y++ {
		for x := 0; x < spec.Width; x++ {
			index := 0
			for plane, offset := range spec.PlaneOffsets {
				bit := data[offset+y*spec.RowStride+x/8] >> uint(7-x%8) & 1
				index |= int(bit) << plane
			}
			if index == 0 && spec.TransparentZero {
				continue
			}
			result.SetNRGBA(x, y, RGB12(spec.Palette[index]))
		}
	}
	return result, nil
}

// Eagle uses the resident copper's four pointers, 44-byte stride and 168 rows.
func Eagle(intro, resident []byte) (*image.NRGBA, error) {
	const paletteOffset = 0x68e - 0x21e
	if len(resident) < paletteOffset+30 {
		return nil, fmt.Errorf("source: missing eagle palette")
	}
	palette := make([]uint16, 16)
	for i := 1; i < len(palette); i++ {
		palette[i] = binary.BigEndian.Uint16(resident[paletteOffset+(i-1)*2:])
	}
	return DecodePlanar(intro, Planar{Width: 352, Height: 168, RowStride: 44,
		PlaneOffsets: []int{0, 0x1ce0, 0x39c0, 0x56a0}, Palette: palette, TransparentZero: true})
}

// SerifFont extracts 59 ASCII cells from the three interleaved original planes.
// The sixtieth storage cell is padding. Advances are independent of cell width.
func SerifFont(resident []byte) (*image.NRGBA, []int, error) {
	const widthOffset, bitmapOffset = 0x17de - 0x21e, 0x19ce - 0x21e
	if len(resident) < bitmapOffset+360*69 {
		return nil, nil, fmt.Errorf("source: missing serif font")
	}
	palette := make([]uint16, 8)
	for i := 1; i < len(palette); i++ {
		palette[i] = binary.BigEndian.Uint16(resident[0x19c0-0x21e+(i-1)*2:])
	}
	atlas, err := DecodePlanar(resident[bitmapOffset:], Planar{Width: 2880, Height: 23, RowStride: 1080,
		PlaneOffsets: []int{0, 360, 720}, Palette: palette, TransparentZero: true})
	if err != nil {
		return nil, nil, err
	}
	advances := make([]int, 59)
	for i := range advances {
		advances[i] = int(binary.BigEndian.Uint16(resident[widthOffset+i*2:])) + 1
	}
	return atlas, advances, nil
}

// MentalTitle decodes the five interleaved planes selected at 0x97d4. The copper
// enables row advancement for 163 rows; indices 1..22 have authored target colors.
func MentalTitle(authors []byte) (*image.NRGBA, error) {
	const start, paletteStart = 0x12058 - 0x9000, 0x99b0 - 0x9000
	if len(authors) < start || len(authors) < paletteStart+44 {
		return nil, fmt.Errorf("source: missing Mental Hangover title")
	}
	palette := make([]uint16, 32)
	for i := 1; i <= 22; i++ {
		palette[i] = binary.BigEndian.Uint16(authors[paletteStart+(i-1)*2:])
	}
	return DecodePlanar(authors[start:], Planar{Width: 352, Height: 163, RowStride: 220,
		PlaneOffsets: []int{0, 44, 88, 132, 176}, Palette: palette, TransparentZero: true})
}
