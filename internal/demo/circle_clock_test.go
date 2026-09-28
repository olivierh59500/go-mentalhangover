package demo

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestOriginalCircleProgramKeepsGlyphContoursAndBoundedProfiles(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/greetings.bin")
	circle, err := source.CirclePart(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(circle.Glyphs) != 59 || len(circle.Glyphs['!'].Contours) != 2 {
		t.Fatal("original outline font is incomplete")
	}
	c := newCircleClock(circle)
	frames := 0
	for c.Step() {
		frames++
		if frames > 10000 {
			t.Fatal("circle program did not terminate")
		}
		if c.active {
			for _, row := range c.rows {
				if c.base+row < 0 || c.base+row > 63 {
					t.Fatal("source curve profile escaped its generated table", c.base, row)
				}
			}
			for _, letter := range c.letters {
				if _, ok := circle.Glyphs[letter]; !ok {
					t.Fatal("circular transport exposed a binary control as a glyph", letter)
				}
			}
		}
	}
	if frames != 3341 {
		t.Fatal("original circle duration changed", frames)
	}
}

func TestCircleControlsAndEveryProfileMatchOriginalMachineCode(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/greetings.bin")
	circle, err := source.CirclePart(data)
	if err != nil {
		t.Fatal(err)
	}
	c := newCircleClock(circle)
	for i := 0; i < 33; i++ {
		c.Step()
	}
	f, err := os.Open("testdata/original-circle-clock.csv")
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
		c.Step()
		if row[0] == "end" {
			if c.active || c.phase != 2 || row[1] != strconv.Itoa(tick) {
				t.Fatal("source circle exit changed", tick)
			}
			break
		}
		got := []int{tick, 0x9f12 + c.cursor, int(c.angle), c.pause, int(c.radiusStep), int(c.factor), int(c.profileStep), int(c.radiusPhase), int(c.profilePhase)}
		for _, value := range c.rows {
			got = append(got, value)
		}
		for i, value := range got {
			want, err := strconv.Atoi(row[i])
			if err != nil || want != value {
				t.Fatalf("circle update %d field %d: Go %d, source %s", tick, i, value, row[i])
			}
		}
	}
}
