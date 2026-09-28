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
		{"sign-raster", 3901, 50}, {"scoopex-demo", 3951, 245},
		{"filled-bobs", 4196, 1971},
		{"filled-vectors", 6167, 245}, {"filled-cube", 6412, 585}, {"faceted-solid", 6997, 676},
		{"stencil-vectors", 7673, 300}, {"gold-arrow", 7973, 318}, {"gold-arrow-pass", 8291, 200},
		{"yellow-cube", 8491, 700}, {"brown-facets", 9191, 800}, {"paired-boxes", 9991, 397},
		{"patterned-pyramid", 10388, 500}, {"hollow-frame", 10888, 950}, {"color-block", 11838, 500},
		{"greetings-load", 12338, 168}, {"star-pages", 12506, 1747},
		{"circle-load", 14253, 100}, {"circle-twist", 14353, 3341},
		{"contact-load", 17694, 100}, {"contact-spheres", 17794, 1194},
		{"perspective-load", 18988, 50}, {"perspective-text", 19038, 1677},
		{"reminder-stars", 20715, 128}, {"always-remember", 20843, 300}, {"finale-load", 21143, 350}, {"checkerboard-finale", 21493, 2000},
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
	if _, _, active := program.At(23493); active {
		t.Fatal("unimplemented next scene counted as completed opening")
	}
}
