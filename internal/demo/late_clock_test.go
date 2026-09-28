package demo

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

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
