package source

import (
	"encoding/binary"
	"fmt"
	"image"
)

type MaskedFace struct {
	Material uint8
	Contours [][]int
}
type PatternedSolid struct {
	Name     string
	Points   []Point3
	Faces    []MaskedFace
	Initial  [6]int16
	Cues     []VectorCue
	Sines    []int16
	Textures []*image.NRGBA
	Period   int
}

// PatternedSolids retains all eight entry calls, including the second arrow
// pass and the two units whose VBL operand is patched to two frames.
func PatternedSolids(data []byte) ([]PatternedSolid, error) {
	s := segment{data, 0x8000}
	sineBank, err := s.read(0xa458, 720)
	if err != nil {
		return nil, err
	}
	sines := make([]int16, 360)
	for i := range sines {
		sines[i] = int16(binary.BigEndian.Uint16(sineBank[i*2:]))
	}
	var solids []PatternedSolid
	for _, d := range []struct {
		name                                    string
		points, faces, motion, palette, texture uint32
		period                                  int
	}{
		{"gold-arrow", 0x9348, 0x9392, 0x93e0, 0x95f8, 0xdcd8, 1},
		{"gold-arrow-pass", 0x9348, 0x9392, 0x9418, 0x95f8, 0xdcd8, 1},
		{"yellow-cube", 0x8f08, 0x8f3a, 0x8f66, 0x95b8, 0xc7d8, 1},
		{"brown-facets", 0x8f9e, 0x9030, 0x90b8, 0x95d8, 0xb2d8, 2},
		{"paired-boxes", 0x90f0, 0x9152, 0x91a8, 0x95c8, 0xd258, 1},
		{"patterned-pyramid", 0x91e0, 0x9200, 0x9222, 0x95e8, 0xbd58, 1},
		{"hollow-frame", 0x9434, 0x94c6, 0x9570, 0x9608, 0xe758, 2},
		{"color-block", 0x925a, 0x92bc, 0x9310, 0x95a8, 0xa858, 1}} {
		model := PatternedSolid{Name: d.name, Sines: sines, Period: d.period}
		n, err := s.word(d.points)
		if err != nil {
			return nil, err
		}
		if n < 3 || n > 128 {
			return nil, fmt.Errorf("source: invalid patterned point count")
		}
		bank, err := s.read(d.points+2, int(n)*6)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(n); i++ {
			model.Points = append(model.Points, Point3{int16(binary.BigEndian.Uint16(bank[i*6:])),
				int16(binary.BigEndian.Uint16(bank[i*6+2:])), int16(binary.BigEndian.Uint16(bank[i*6+4:]))})
		}
		n, err = s.word(d.faces)
		if err != nil {
			return nil, err
		}
		if n < 1 || n > 256 {
			return nil, fmt.Errorf("source: invalid patterned face count")
		}
		address := d.faces + 2
		for i := 0; i < int(n); i++ {
			header, err := s.read(address, 2)
			if err != nil {
				return nil, err
			}
			if header[0] < 1 || header[0] > 7 || header[1] < 3 {
				return nil, fmt.Errorf("source: invalid patterned face")
			}
			indices, err := s.read(address+2, int(header[1])+1)
			if err != nil {
				return nil, err
			}
			face := MaskedFace{Material: header[0]}
			contour := []int{}
			finish := func() error {
				if len(contour) < 4 || contour[0] != contour[len(contour)-1] {
					return fmt.Errorf("source: open patterned contour")
				}
				face.Contours = append(face.Contours, contour)
				contour = nil
				return nil
			}
			for _, index := range indices {
				if index >= 128 {
					if err := finish(); err != nil {
						return nil, err
					}
					continue
				}
				if int(index) >= len(model.Points) {
					return nil, fmt.Errorf("source: missing patterned vertex")
				}
				contour = append(contour, int(index))
			}
			if err := finish(); err != nil {
				return nil, err
			}
			model.Faces = append(model.Faces, face)
			address += uint32(header[1]) + 3
		}
		model.Initial, model.Cues, err = ReadMotion(data, 0x8000, d.motion)
		if err != nil {
			return nil, err
		}
		bank, err = s.read(d.palette, 16)
		if err != nil {
			return nil, err
		}
		palette := make([]uint16, 8)
		for i := range palette {
			palette[i] = binary.BigEndian.Uint16(bank[i*2:])
		}
		bank, err = s.read(d.texture, 28*96)
		if err != nil {
			return nil, err
		}
		atlas, err := DecodePlanar(bank, Planar{Width: 224, Height: 32, RowStride: 84, PlaneOffsets: []int{0, 28, 56}, Palette: palette, TransparentZero: true})
		if err != nil {
			return nil, err
		}
		for i := 0; i < 7; i++ {
			texture := image.NewNRGBA(image.Rect(0, 0, 256, 256))
			for y := 0; y < 256; y++ {
				for x := 0; x < 256; x++ {
					texture.SetNRGBA(x, y, atlas.NRGBAAt(i*32+x%32, y%32))
				}
			}
			model.Textures = append(model.Textures, texture)
		}
		solids = append(solids, model)
	}
	return solids, nil
}
