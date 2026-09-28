package demo

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestPerspectivePointPlanesMatchOriginalORRaster(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/final-reminder.bin")
	data, err := source.PerspectivePart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newPerspectiveClock(data)
	for _, row := range originalLateRows(t, "perspective-points") {
		frame := lateInt(t, row[0])
		c.offset = [3]int16{int16(-6 * frame), int16(3 * frame), int16(-10 * frame)}
		c.rasterizePoints()
		planes := make([]byte, 88*200)
		for _, index := range c.touched {
			x, y := index%Width, index/Width
			for plane := 0; plane < 2; plane++ {
				if c.pointMasks[index]>>uint(plane)&1 != 0 {
					planes[y*88+plane*44+x/8] |= 1 << uint(7-x%8)
				}
			}
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(planes)); got != row[1] {
			t.Fatalf("perspective point frame %d differs from original bitplanes: %s", frame, got)
		}
	}
}

func originalLateRows(t *testing.T, name string) [][]string {
	t.Helper()
	f, err := os.Open("testdata/original-" + name + ".csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func lateInt(t *testing.T, value string) int {
	t.Helper()
	i, err := strconv.Atoi(value)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

// Fixtures execute the original matrix and blitter setup, without GPU emulation.
func TestContactMotionAndVisiblePosesMatchOriginalCPU(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/circle-scroll.bin")
	data, err := source.ContactPart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newContactClock(data)
	poses := originalLateRows(t, "contact-poses")
	poseIndex := 0
	rows := originalLateRows(t, "contact-clock")
	if len(rows) != 1194 {
		t.Fatal("incomplete original motion fixture")
	}
	for tick, row := range rows {
		if !c.Step() {
			t.Fatal("contact stopped early", tick)
		}
		got := []int{tick, int(c.angles[0]), int(c.angles[1]), int(c.angles[2]), int(c.offset[0]), int(c.offset[1]), c.head, c.depth}
		for i, value := range got {
			if value != lateInt(t, row[i]) {
				t.Fatalf("contact tick %d field %d: Go %d, CPU %s", tick, i, value, row[i])
			}
		}
		for poseIndex < len(poses) && lateInt(t, poses[poseIndex][0]) == tick {
			row := poses[poseIndex]
			index := lateInt(t, row[1])
			p := c.poses[index]
			if !p.visible || p.x != lateInt(t, row[2]) || p.y+26 != lateInt(t, row[3]) || p.frame != lateInt(t, row[4]) || p.height != lateInt(t, row[5]) {
				t.Fatalf("contact tick %d sphere %d: Go %+v, CPU %v", tick, index, p, row)
			}
			poseIndex++
		}
	}
	if poseIndex != len(poses) {
		t.Fatal("not all original poses compared")
	}
}

func TestPerspectiveLookupAndCompleteTransportMatchOriginalCPU(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/final-reminder.bin")
	data, err := source.PerspectivePart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newPerspectiveClock(data)
	table := originalLateRows(t, "perspective-table")
	if len(table) != 3080 {
		t.Fatal("incomplete original lookup fixture")
	}
	for _, row := range table {
		r, col := lateInt(t, row[0]), lateInt(t, row[1])
		p := c.Point(col/22, source.PolarGlyphPoint{Angle: byte(col % 22), Row: byte(r)})
		if int(p.X) != lateInt(t, row[2]) || int(p.Y)-8 != lateInt(t, row[3]) {
			t.Fatalf("perspective row %d col %d: Go %+v, CPU %v", r, col, p, row)
		}
	}
	for i := 0; i < 88; i++ {
		c.Step()
	}
	for tick, row := range originalLateRows(t, "perspective-clock") {
		c.Step()
		if row[0] == "end" {
			if lateInt(t, row[1]) != tick || tick != 1507 || c.phase != 3 {
				t.Fatal("perspective original ending changed", tick)
			}
			break
		}
		if c.cursor != lateInt(t, row[1]) || int(c.position) != lateInt(t, row[2]) {
			t.Fatalf("perspective tick %d: Go %d,%d; CPU %v", tick, c.cursor, c.position, row)
		}
	}
}

func TestContactSphereProgramRetainsSizeAndProjectionBounds(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/circle-scroll.bin")
	data, err := source.ContactPart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Points) != 130 || len(data.Sines) != 360 || data.Title.Bounds().Dy() != 272 {
		t.Fatal("contact banks incomplete")
	}
	c := newContactClock(data)
	frames := 0
	for c.Step() {
		frames++
		for _, p := range c.poses {
			if p.frame < 0 || p.frame > 15 || p.height != 16-p.frame {
				t.Fatal("original sphere size bank escaped")
			}
		}
	}
	if frames != 1194 {
		t.Fatal("contact operation counts changed", frames)
	}
}

func TestPerspectiveMessageAndOutlineLookupCompleteWithoutCutoff(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/final-reminder.bin")
	data, err := source.PerspectivePart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newPerspectiveClock(data)
	frames := 0
	for c.Step() {
		frames++
		if frames > 5000 {
			t.Fatal("perspective message never ended")
		}
	}
	if frames != 1677 || len(data.Text) != 146 {
		t.Fatal("perspective message duration changed", frames, len(data.Text))
	}
}
