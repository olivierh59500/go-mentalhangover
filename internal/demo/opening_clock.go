package demo

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/timeline"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

const openingMusicTick = 565

type openingUnit struct {
	Name          string
	Kind          string
	Start, Frames int
	Card, Model   int
	Hold          int
}

// openingProgram keeps source-defined operation counts. The boot/load holds
// precede the music and are anchored to the supplied recording's visible timing.
func openingProgram(models []source.VectorModel) ([]openingUnit, *timeline.CueRanges, error) {
	units := []openingUnit{{Name: "startup", Kind: "blank", Start: 0, Frames: 118},
		{Name: "eagle", Kind: "eagle", Start: 118, Frames: openingMusicTick + 16 - 118}}
	appendUnit := func(name, kind string, frames, card, model, hold int) {
		last := units[len(units)-1]
		units = append(units, openingUnit{Name: name, Kind: kind, Start: last.Start + last.Frames,
			Frames: frames, Card: card, Model: model, Hold: hold})
	}
	for i, name := range []string{"presents", "new-demonstration", "called"} {
		appendUnit(name, "card", 166, i, 0, 101)
	}
	appendUnit("mental-title", "title", 256, 0, 0, 191)
	for i, name := range []string{"coded-by", "graphics-by", "musics-by", "follow-sign"} {
		appendUnit(name, "card", 166, i+3, 0, 101)
		if i >= len(models) {
			return nil, nil, fmt.Errorf("original opening vector bank is incomplete")
		}
		frames := 0
		for _, cue := range models[i].Cues {
			frames += cue.Frames
		}
		appendUnit(models[i].Name, "vector", frames, 0, i, 0)
	}
	appendUnit("sign-raster", "sign-raster", 50, 0, 3, 0)
	// This interlude includes a disk-load hold. Its duration is anchored to the
	// supplied recording; the flash and fade keep their original frame counts.
	appendUnit("scoopex-demo", "card", 245, 7, 0, 180)
	appendUnit("filled-bobs", "bobs", 1971, 0, 0, 0)
	appendUnit("filled-vectors", "card", 245, 8, 0, 180)
	appendUnit("filled-cube", "solid", 585, 0, 0, 0)
	appendUnit("faceted-solid", "solid", 676, 0, 1, 0)
	appendUnit("stencil-vectors", "card", 300, 9, 0, 235)
	for i, definition := range []struct {
		name   string
		frames int
	}{
		{"gold-arrow", 318}, {"gold-arrow-pass", 200}, {"yellow-cube", 700}, {"brown-facets", 800},
		{"paired-boxes", 397}, {"patterned-pyramid", 500}, {"hollow-frame", 950}, {"color-block", 500}} {
		appendUnit(definition.name, "pattern", definition.frames, 0, i, 0)
	}
	appendUnit("greetings-load", "blank", 168, 0, 0, 0)
	appendUnit("star-pages", "star-pages", 1747, 0, 0, 0)
	appendUnit("circle-load", "blank", 100, 0, 0, 0)
	appendUnit("circle-twist", "circle", 3341, 0, 0, 0)
	appendUnit("contact-load", "blank", 100, 0, 0, 0)
	appendUnit("contact-spheres", "contact", 1194, 0, 0, 0)
	appendUnit("perspective-load", "blank", 50, 0, 0, 0)
	appendUnit("perspective-text", "perspective", 1677, 0, 0, 0)
	appendUnit("reminder-stars", "blank", 128, 0, 0, 0)
	appendUnit("always-remember", "card", 300, 10, 0, 235)
	appendUnit("finale-load", "blank", 350, 0, 0, 0)
	appendUnit("checkerboard-finale", "finale", 2000, 0, 0, 0)
	ranges := make([]timeline.CueRange, len(units))
	for i, unit := range units {
		ranges[i] = timeline.CueRange{Start: float64(unit.Start), End: float64(unit.Start + unit.Frames)}
	}
	program, err := timeline.NewCueRanges(ranges)
	return units, program, err
}
