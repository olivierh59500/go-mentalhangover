package demo

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestFilledSolidProjectionMatchesOriginalCPUIncludingCueBoundaries(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/stencil.bin")
	solids, err := source.FilledSolids(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("testdata/original-solid-projection.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		solid := solids[0]
		if strings.HasPrefix(row[0], "faceted-solid-") {
			solid = solids[1]
		}
		var state [6]int16
		for i := range state {
			value, err := strconv.ParseInt(row[i+1], 10, 16)
			if err != nil {
				t.Fatal(err)
			}
			state[i] = int16(value)
		}
		want, err := hex.DecodeString(row[7])
		if err != nil || len(want) != len(solid.Model.Points)*4 {
			t.Fatal("invalid filled-solid fixture", row[0])
		}
		got := make([]source.Point2, len(solid.Model.Points))
		if err := projectSolid(solid.Model.Points, solidMatrix([3]int16{state[0], state[1], state[2]}, solid.Sines), state, got); err != nil {
			t.Fatal(err)
		}
		for i, p := range got {
			x, y := int16(binary.BigEndian.Uint16(want[i*4:])), int16(binary.BigEndian.Uint16(want[i*4+2:]))
			if p.X != x || p.Y != y {
				t.Fatalf("%s point %d: Go %v, original (%d,%d)", row[0], i, p, x, y)
			}
		}
		clock := newVectorClock(source.VectorModel{Initial: solid.Initial, Cues: solid.Cues})
		tick, err := strconv.Atoi(row[0][strings.LastIndexByte(row[0], '-')+1:])
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= tick; i++ {
			if !clock.Step() {
				t.Fatal("solid cue bank ended early")
			}
		}
		if clock.state != state {
			t.Fatal("filled solid retained the wrong boundary pose", row[0], clock.state, state)
		}
	}
}
