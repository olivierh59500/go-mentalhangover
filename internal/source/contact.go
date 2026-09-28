package source

import (
	"encoding/binary"
	"image"
)

type ContactData struct {
	Points []Point2
	Sines  []int16
	Balls  *image.NRGBA
	Title  *image.NRGBA
}

// ContactPart preserves all 130 source points, the sixteen sphere sizes and
// the eight hardware-sprite columns of the contact text from cylinder 46.
func ContactPart(data []byte) (ContactData, error) {
	var result ContactData
	s := segment{data, 0x9000}
	bank, err := s.read(0xbd96, 130*4)
	if err != nil {
		return result, err
	}
	for i := 0; i < 130; i++ {
		result.Points = append(result.Points, Point2{int16(binary.BigEndian.Uint16(bank[i*4:])), int16(binary.BigEndian.Uint16(bank[i*4+2:]))})
	}
	bank, err = s.read(0xc1a6, 720)
	if err != nil {
		return result, err
	}
	for i := 0; i < 360; i++ {
		result.Sines = append(result.Sines, int16(binary.BigEndian.Uint16(bank[i*2:])))
	}
	bank, err = s.read(0xb996, 64*16)
	if err != nil {
		return result, err
	}
	result.Balls, err = DecodePlanar(bank, Planar{Width: 256, Height: 16, RowStride: 64, PlaneOffsets: []int{0, 32}, Palette: []uint16{0, 0xfff, 0x34b, 0x017}})
	if err != nil {
		return result, err
	}
	maskBank, err := s.read(0xb596, 64*16)
	if err != nil {
		return result, err
	}
	mask, err := DecodePlanar(maskBank, Planar{Width: 256, Height: 16, RowStride: 64, PlaneOffsets: []int{0, 32}, Palette: []uint16{0, 0xfff, 0xfff, 0xfff}, TransparentZero: true})
	if err != nil {
		return result, err
	}
	for i := 3; i < len(result.Balls.Pix); i += 4 {
		result.Balls.Pix[i] = mask.Pix[i]
	}
	result.Title = image.NewNRGBA(image.Rect(0, 0, 352, 272))
	for i := 0; i < 8; i++ {
		address := uint32(0x9b76 + i*0x344)
		header, err := s.read(address, 4)
		if err != nil {
			return result, err
		}
		pos, ctl := binary.BigEndian.Uint16(header), binary.BigEndian.Uint16(header[2:])
		start := int(pos>>8) + int(ctl>>2&1)*256
		end := int(ctl>>8) + int(ctl>>1&1)*256
		x0, y0 := int(pos&255)*2+int(ctl&1)-113, start-34
		bank, err := s.read(address+4, (end-start)*4)
		if err != nil {
			return result, err
		}
		column, err := DecodePlanar(bank, Planar{Width: 16, Height: end - start, RowStride: 4, PlaneOffsets: []int{0, 2}, Palette: []uint16{0, 0xdef, 0x7bc, 0x168}, TransparentZero: true})
		if err != nil {
			return result, err
		}
		for y := 0; y < end-start; y++ {
			for x := 0; x < 16; x++ {
				result.Title.SetNRGBA(x0+x, y0+y, column.NRGBAAt(x, y))
			}
		}
	}
	return result, nil
}
