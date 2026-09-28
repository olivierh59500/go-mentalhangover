//go:build mental_vector_rendercheck

package demo

import (
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
