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

func TestPatternedProjectionMatchesOriginalCPUForAllEightObjects(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/patterned-vectors.bin")
	models, err := source.PatternedSolids(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("testdata/original-pattern-projection.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		name := row[0][:strings.LastIndexByte(row[0], '-')]
		var model source.PatternedSolid
		for _, m := range models {
			if m.Name == name {
				model = m
			}
		}
		if len(model.Points) == 0 {
			t.Fatal("fixture model missing", name)
		}
		var state [6]int16
		for i := range state {
			v, err := strconv.ParseInt(row[i+1], 10, 16)
			if err != nil {
				t.Fatal(err)
			}
			state[i] = int16(v)
		}
		want, err := hex.DecodeString(row[7])
		if err != nil || len(want) != len(model.Points)*4 {
			t.Fatal("invalid stencil projection fixture", row[0])
		}
		got := make([]source.Point2, len(model.Points))
		if err := projectPatterned(model.Points, patternedMatrix([3]int16{state[0], state[1], state[2]}, model.Sines), state, got); err != nil {
			t.Fatal(err)
		}
		for i, p := range got {
			x, y := int16(binary.BigEndian.Uint16(want[i*4:])), int16(binary.BigEndian.Uint16(want[i*4+2:]))
			if p.X != x || p.Y != y {
				t.Fatalf("%s point %d: Go %v, original (%d,%d)", row[0], i, p, x, y)
			}
		}
	}
}
