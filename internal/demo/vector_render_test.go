//go:build mental_vector_rendercheck

package demo

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestMain(m *testing.M) { os.Exit(testutil.RunGPU(m)) }

func TestVectorBatchUsesOneEvenOddFillForConcaveContoursAndHoles(t *testing.T) {
	// The outer L shape and inner rectangle share the source's single mask.
	model := source.VectorModel{Points: []source.Point2{{2, 2}, {22, 2}, {22, 8}, {10, 8}, {10, 22}, {2, 22},
		{4, 4}, {7, 4}, {7, 7}, {4, 7}},
		Contours: []source.Contour{{Material: 1, Indices: []int{0, 1, 2, 3, 4, 5, 0}},
			{Material: 1, Indices: []int{6, 7, 8, 9, 6}}},
		Cues: []source.VectorCue{{Frames: 1}}}
	effect, err := newVectorEffect(model, make([]int16, 2048))
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	copy(effect.points, model.Points)
	effect.color = color.NRGBA{B: 255, A: 255}
	effect.ready = true
	dst := ebiten.NewImage(24, 24)
	defer dst.Deallocate()
	effect.Draw(dst)
	pixels := image.NewRGBA(image.Rect(0, 0, 24, 24))
	dst.ReadPixels(pixels.Pix)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			outer := x >= 2 && x < 22 && y >= 2 && y < 8 || x >= 2 && x < 10 && y >= 8 && y < 22
			hole := x >= 4 && x < 7 && y >= 4 && y < 7
			want := outer && !hole
			got := pixels.RGBAAt(x, y).A != 0
			if got != want {
				t.Fatalf("parity mask differs at %d,%d: want %t, got %t", x, y, want, got)
			}
		}
	}
}

func TestCopperShaderMatchesAllOriginalNibbles(t *testing.T) {
	src, dst := ebiten.NewImage(64, 64), ebiten.NewImage(64, 64)
	defer src.Deallocate()
	defer dst.Deallocate()
	pixels := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for i := 1; i < 4096; i++ {
		pixels.SetNRGBA(i%64, i/64, source.RGB12(uint16(i)))
	}
	src.WritePixels(pixels.Pix)
	palette, err := newCopperPalette()
	if err != nil {
		t.Fatal(err)
	}
	defer palette.Close()
	for _, mode := range []source.PaletteMode{source.PaletteGray, source.PaletteFromWhite, source.PaletteToBlack} {
		for _, level := range []int{0, 1, 7, 15, 16, 31, 32} {
			if mode != source.PaletteFromWhite && level > 16 {
				continue
			}
			dst.Clear()
			palette.Draw(dst, src, mode, level)
			got := make([]byte, len(pixels.Pix))
			dst.ReadPixels(got)
			for i := 0; i < 4096; i++ {
				want := color.NRGBA{}
				if i != 0 {
					want = source.RGB12(source.PaletteWord(uint16(i), mode, level))
				}
				if !bytes.Equal(got[i*4:i*4+4], []byte{want.R, want.G, want.B, want.A}) {
					t.Fatalf("palette mode %d level %d color %03x: %v != %v", mode, level, i, got[i*4:i*4+4], want)
				}
			}
		}
	}
}

func TestOpeningRepeatedDrawPreservesClockAndVectorPose(t *testing.T) {
	game, err := NewOpening(true)
	if err != nil {
		t.Fatal(err)
	}
	defer game.Close()
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for tick := 0; tick < 1700; tick++ {
		if tick > 0 {
			if err := game.Update(); err != nil {
				t.Fatal(err)
			}
		}
		if tick%137 != 0 {
			continue
		}
		clock, local := game.clock.Tick(), game.localTick
		state := game.vectors[0].clock.state
		game.Draw(dst)
		a := make([]byte, Width*Height*4)
		dst.ReadPixels(a)
		game.Draw(dst)
		b := make([]byte, len(a))
		dst.ReadPixels(b)
		if clock != game.clock.Tick() || local != game.localTick || state != game.vectors[0].clock.state || !bytes.Equal(a, b) {
			t.Fatal("drawing changes source timeline or pixels", tick)
		}
	}
}
