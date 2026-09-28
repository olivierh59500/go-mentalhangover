package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestFilledSolidBanksPreserveModelsAndSourceCueCounts(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/stencil.bin")
	solids, err := source.FilledSolids(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(solids) != 2 || len(solids[0].Model.Points) != 8 || len(solids[1].Model.Points) != 20 || len(solids[1].Model.Faces) != 24 {
		t.Fatal("original filled solids missing")
	}
	for i, solid := range solids {
		frames := 0
		for _, cue := range solid.Cues {
			frames += cue.Frames
		}
		if frames != []int{585, 676}[i] {
			t.Fatal("filled vector cue duration changed", i, frames)
		}
		if len(solid.Sines) != 2048 {
			t.Fatal("sine bank incomplete")
		}
		for j := 0; j < 512; j++ {
			if solid.Sines[j] != solid.Sines[1023-j] || solid.Sines[j] != -solid.Sines[j+1024] {
				t.Fatal("original sine mirroring changed", j)
			}
		}
	}
}
