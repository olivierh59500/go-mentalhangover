package source

import (
	"encoding/binary"
	"fmt"
)

// ReadOutlineFont decodes the shared byte-pair and separated-contour font format.
// The consumer maps its first coordinate to a polar angle or perspective column.
func ReadOutlineFont(data []byte, base, table uint32, maxRow int) (map[byte]PolarGlyph, error) {
	s := segment{data, base}
	glyphs := make(map[byte]PolarGlyph, 59)
	for i := 0; i < 59; i++ {
		pointer, err := s.read(table+uint32(i*4), 4)
		if err != nil {
			return nil, err
		}
		address := binary.BigEndian.Uint32(pointer)
		header, err := s.read(address, 1)
		if err != nil {
			return nil, err
		}
		count := int(header[0])
		address++
		glyph := PolarGlyph{}
		if count == 0 {
			glyphs[byte(i+32)] = glyph
			continue
		}
		var contour []PolarGlyphPoint
		readPoint := func() (PolarGlyphPoint, error) {
			bank, err := s.read(address, 2)
			if err != nil {
				return PolarGlyphPoint{}, err
			}
			address += 2
			if bank[0] >= 128 || int(bank[1]) > maxRow {
				return PolarGlyphPoint{}, fmt.Errorf("source: invalid outline-font vertex")
			}
			return PolarGlyphPoint{bank[0], bank[1]}, nil
		}
		p, err := readPoint()
		if err != nil {
			return nil, err
		}
		contour = append(contour, p)
		for edge := 0; edge < count; {
			marker, err := s.read(address, 1)
			if err != nil {
				return nil, err
			}
			if marker[0] >= 128 {
				glyph.Contours = append(glyph.Contours, contour)
				contour = nil
				address++
				p, err := readPoint()
				if err != nil {
					return nil, err
				}
				contour = append(contour, p)
				continue
			}
			p, err := readPoint()
			if err != nil {
				return nil, err
			}
			contour = append(contour, p)
			edge++
		}
		glyph.Contours = append(glyph.Contours, contour)
		for _, c := range glyph.Contours {
			if len(c) < 4 || c[0] != c[len(c)-1] {
				return nil, fmt.Errorf("source: open outline-font contour for %q", byte(i+32))
			}
		}
		glyphs[byte(i+32)] = glyph
	}
	return glyphs, nil
}
