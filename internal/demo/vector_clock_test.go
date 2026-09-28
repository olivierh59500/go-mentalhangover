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

// These coordinates came from executing the original 68000 machine code in
// Ghidra's existing CPU emulator, not from the translated Go matrix formulas.
func TestProjectionMatchesOriginalMachineCode(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/authors-ribbon.bin")
	models, err := source.AuthorModels(data)
	if err != nil {
		t.Fatal(err)
	}
	sines, _ := source.AuthorSines(data)
	fixture, err := os.Open("testdata/original-projection.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Close()
	rows, err := csv.NewReader(fixture).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		name := row[0][:strings.LastIndexByte(row[0], '-')]
		var model source.VectorModel
		for _, candidate := range models {
			if candidate.Name == name {
				model = candidate
			}
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
		if err != nil || len(want) != len(model.Points)*4 {
			t.Fatal("invalid original projection fixture", row[0], err)
		}
		got := make([]source.Point2, len(model.Points))
		if err := projectVector(model.Points, vectorMatrix(state, sines), state, got); err != nil {
			t.Fatal(err)
		}
		for i, point := range got {
			x, y := int16(binary.BigEndian.Uint16(want[i*4:])), int16(binary.BigEndian.Uint16(want[i*4+2:]))
			if point.X != x || point.Y != y {
				t.Fatalf("%s vertex %d: Go (%d,%d), original (%d,%d)", row[0], i, point.X, point.Y, x, y)
			}
		}
	}
}

func TestVectorCueBoundaryRetainsPoseAndUsesNewDelta(t *testing.T) {
	model := source.VectorModel{Initial: [6]int16{32766, 0, 0, 0, 0, 0},
		Cues: []source.VectorCue{{Frames: 2, Delta: [6]int16{2, 0, 0, 0, 0, 1}},
			{Frames: 3, Delta: [6]int16{-1, 0, 0, 0, 0, -2}}}}
	clock := newVectorClock(model)
	expected := [][2]int16{{-32768, 1}, {-32766, 2}, {-32767, 0}, {-32768, -2}, {32767, -4}}
	for i, want := range expected {
		if !clock.Step() || clock.state[0] != want[0] || clock.state[5] != want[1] {
			t.Fatal("word overflow or cue-boundary phase changed", i, clock.state)
		}
	}
	last := clock.state
	if clock.Step() || clock.Step() || clock.state != last {
		t.Fatal("completed cue bank advanced or reset the final pose")
	}
}

func TestOriginalVectorProgramHasBoundedProjectionAndNoStepAllocations(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/authors-ribbon.bin")
	models, err := source.AuthorModels(data)
	if err != nil {
		t.Fatal(err)
	}
	sines, err := source.AuthorSines(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range models {
		clock := newVectorClock(model)
		points := make([]source.Point2, len(model.Points))
		steps := 0
		for clock.Step() {
			if err := projectVector(model.Points, vectorMatrix(clock.state, sines), clock.state, points); err != nil {
				t.Fatal(model.Name, steps, err)
			}
			steps++
		}
		expected := 0
		for _, cue := range model.Cues {
			expected += cue.Frames
		}
		if steps != expected {
			t.Fatal("vector runtime differs from original cue lengths", model.Name, steps, expected)
		}
		clock = newVectorClock(model)
		if allocations := testing.AllocsPerRun(100, func() {
			clock.Step()
			_ = projectVector(model.Points, vectorMatrix(clock.state, sines), clock.state, points)
		}); allocations != 0 {
			t.Fatal("original vector projection allocated", model.Name, allocations)
		}
	}
}
