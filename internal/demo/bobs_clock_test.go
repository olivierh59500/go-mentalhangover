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

func originalBOBs(t *testing.T) (source.BOBData, []source.SolidModel) {
	t.Helper()
	data, _ := assets.Files.ReadFile("raw/filled-vector.bin")
	bobs, err := source.ReadBOBs(data)
	if err != nil {
		t.Fatal(err)
	}
	models, err := source.BOBModels(data)
	if err != nil {
		t.Fatal(err)
	}
	return bobs, models
}

// Every text update is checked against the unmodified 0x98a0 CPU routine.
// This fixture includes acceleration, look-ahead controls and paused transport.
func TestBOBClockMatchesOriginalCPUForCompleteMessage(t *testing.T) {
	bobs, _ := originalBOBs(t)
	c := newBOBClock(bobs)
	f, err := os.Open("testdata/original-bob-clock.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for tick, row := range rows {
		if row[0] == "end" {
			if row[1] != strconv.Itoa(tick) || tick != 1971 || c.Step() {
				t.Fatal("original BOB ending changed", tick)
			}
			break
		}
		if !c.Step() {
			t.Fatal("BOB program stopped early", tick)
		}
		cursor := uint32(0xaed6)
		if c.cursor < len(bobs.Tokens) {
			cursor = bobs.Tokens[c.cursor].Address
		}
		got := []int{tick, int(cursor), c.remaining, c.speed, c.targetSpeed, c.pause, int(c.target), c.count,
			int(c.stepX), int(c.stepY), int(c.gapX), int(c.gapY), int(c.rotation[0]), int(c.rotation[1]), int(c.rotation[2])}
		for i, value := range got {
			want, err := strconv.Atoi(row[i])
			if err != nil || value != want {
				t.Fatalf("BOB update %d field %d: Go %d, original %s (%v)", tick, i, value, row[i], err)
			}
		}
	}
}

func TestBOBProjectionMatchesOriginal68000MatrixAndDivision(t *testing.T) {
	bobs, models := originalBOBs(t)
	f, err := os.Open("testdata/original-bob-projection.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		model := models[0]
		if strings.HasPrefix(row[0], "pyramid-") {
			model = models[1]
		}
		var angles [3]int16
		for i := range angles {
			value, err := strconv.ParseInt(row[i+1], 10, 16)
			if err != nil {
				t.Fatal(err)
			}
			angles[i] = int16(value)
		}
		depth, err := strconv.ParseInt(row[4], 10, 16)
		if err != nil {
			t.Fatal(err)
		}
		want, err := hex.DecodeString(row[5])
		if err != nil || len(want) != len(model.Points)*4 {
			t.Fatal("invalid original BOB fixture", row[0])
		}
		got := make([]source.Point2, len(model.Points))
		if err := projectBOB(model.Points, solidMatrix(angles, bobs.Sines), int16(depth), got); err != nil {
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

func TestBOBSteppingAndAllOriginalCopyOffsetsAreBounded(t *testing.T) {
	bobs, models := originalBOBs(t)
	c := newBOBClock(bobs)
	points := make([]source.Point2, 8)
	for tick := 0; c.Step(); tick++ {
		for i := 0; i < c.count; i++ {
			x, y := c.Offset(i)
			if x < -16 || x > 304 || y < 34 || y > 201 {
				t.Fatal("source lookup escaped its original bounds", tick, x, y)
			}
		}
		if c.depth != 4100 {
			model := models[c.model]
			if err := projectBOB(model.Points, solidMatrix(c.angles, bobs.Sines), c.depth, points[:len(model.Points)]); err != nil {
				t.Fatal(err)
			}
			for _, p := range points[:len(model.Points)] {
				if p.X < -32 || p.X >= 64 || p.Y < -32 || p.Y >= 64 {
					t.Fatal("original BOB escaped its guarded source area", tick, p)
				}
			}
		}
	}
	c = newBOBClock(bobs)
	if got := testing.AllocsPerRun(100, func() { c.Step() }); got != 0 {
		t.Fatal("BOB clocks allocate during playback", got)
	}
}
