package source

import (
	"encoding/binary"
	"fmt"
	"image"
)

// SignRaster decodes the colors consumed by the final credit's 27 blitter
// copies. Each word colors ten copper rows; the source advances two words per
// frame for fifty frames. Copper row 0x20 lies two rows above the visible area.
func SignRaster(authors []byte) (*image.NRGBA, error) {
	const offset, words = 0xa862 - 0x9000, 126
	if len(authors) < offset+words*2 {
		return nil, fmt.Errorf("source: sign raster bank is truncated")
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 1, words*10))
	for i := 0; i < words; i++ {
		word := binary.BigEndian.Uint16(authors[offset+i*2:])
		for y := 0; y < 10; y++ {
			pixels.SetNRGBA(0, i*10+y, RGB12(word))
		}
	}
	return pixels, nil
}
