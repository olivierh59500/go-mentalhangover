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
	ranges := make([]timeline.CueRange, len(units))
	for i, unit := range units {
		ranges[i] = timeline.CueRange{Start: float64(unit.Start), End: float64(unit.Start + unit.Frames)}
	}
	program, err := timeline.NewCueRanges(ranges)
	return units, program, err
}
