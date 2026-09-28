package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestPatternedUnitsKeepAllEightCallsAndPatchedCadences(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/patterned-vectors.bin")
	solids, err := source.PatternedSolids(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(solids) != 8 {
		t.Fatal("patterned production calls missing")
	}
	total := 0
	for i, solid := range solids {
		frames := 0
		for _, cue := range solid.Cues {
			frames += cue.Frames * solid.Period
		}
		if frames != []int{318, 200, 700, 800, 397, 500, 950, 500}[i] {
			t.Fatal("original stencil clock changed", solid.Name, frames)
		}
		total += frames
		for _, face := range solid.Faces {
			for _, contour := range face.Contours {
				if contour[0] != contour[len(contour)-1] {
					t.Fatal("open original contour")
				}
			}
		}
		if len(solid.Textures) != 7 || len(solid.Sines) != 360 {
			t.Fatal("original pattern materials missing")
		}
	}
	if total != 4365 {
		t.Fatal("total stencil duration changed", total)
	}
}
