package source

import "testing"

func TestInterleavedBitplanesKeepBitOrderAndTransparency(t *testing.T) {
	data := []byte{0xa0, 0x60, 0x20, 0x10}
	image, err := DecodePlanar(data, Planar{Width: 4, Height: 2, RowStride: 2,
		PlaneOffsets: []int{0, 1}, Palette: []uint16{0, 0xf00, 0x0f0, 0x00f}, TransparentZero: true})
	if err != nil {
		t.Fatal(err)
	}
	indices := [][]uint16{{0xf00, 0x0f0, 0x00f, 0}, {0, 0, 0xf00, 0x0f0}}
	for y, row := range indices {
		for x, value := range row {
			got := image.NRGBAAt(x, y)
			if value == 0 {
				if got.A != 0 {
					t.Fatal("zero bitplane index lost transparency")
				}
			} else if got != RGB12(value) {
				t.Fatal("bit significance or interleaved row order changed", x, y, got)
			}
		}
	}
	if _, err := DecodePlanar(data[:3], Planar{Width: 4, Height: 2, RowStride: 2,
		PlaneOffsets: []int{0, 1}, Palette: []uint16{0, 1, 2, 3}}); err == nil {
		t.Fatal("missing last plane row accepted")
	}
}
