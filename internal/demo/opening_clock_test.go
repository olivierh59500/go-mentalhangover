package demo

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestOpeningRetainsOriginalOperationCountsAndBoundaryOrder(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/authors-ribbon.bin")
	models, err := source.AuthorModels(data)
	if err != nil {
		t.Fatal(err)
	}
	units, program, err := openingProgram(models)
	if err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		name          string
		start, frames int
	}{
		{"startup", 0, 118}, {"eagle", 118, 463},
		{"presents", 581, 166}, {"new-demonstration", 747, 166}, {"called", 913, 166},
		{"mental-title", 1079, 256}, {"coded-by", 1335, 166}, {"slayer", 1501, 532},
		{"graphics-by", 2033, 166}, {"reward", 2199, 570},
		{"musics-by", 2769, 166}, {"uncle-tom", 2935, 544},
		{"follow-sign", 3479, 166}, {"sign", 3645, 256},
	}
	if len(units) != len(expected) {
		t.Fatal("opening units missing")
	}
	for i, want := range expected {
		unit := units[i]
		if unit.Name != want.name || unit.Start != want.start || unit.Frames != want.frames {
			t.Fatal("source sequence changed", unit, want)
		}
		for _, tick := range []int{unit.Start, unit.Start + unit.Frames - 1} {
			index, _, active := program.At(float64(tick))
			if !active || index != i {
				t.Fatal("wrong boundary scene", tick, index)
			}
		}
	}
	if _, _, active := program.At(3901); active {
		t.Fatal("unimplemented next scene counted as completed opening")
	}
}
