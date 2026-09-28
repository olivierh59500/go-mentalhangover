package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestOriginalCreditBanksKeepPackedContoursAndCueDurations(t *testing.T) {
	data, err := assets.Files.ReadFile("raw/authors-ribbon.bin")
	if err != nil {
		t.Fatal(err)
	}
	models, err := source.AuthorModels(data)
	if err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		name                 string
		points, faces, ticks int
	}{
		{"slayer", 57, 8, 532},
		{"reward", 76, 10, 570},
		{"uncle-tom", 48, 9, 544},
		{"sign", 53, 3, 256},
	}
	for i, model := range models {
		want := expected[i]
		ticks := 0
		for _, cue := range model.Cues {
			ticks += cue.Frames
		}
		if model.Name != want.name || len(model.Points) != want.points || len(model.Contours) != want.faces || ticks != want.ticks {
			t.Fatalf("authored bank changed: %s points=%d faces=%d ticks=%d", model.Name, len(model.Points), len(model.Contours), ticks)
		}
		for _, contour := range model.Contours {
			if contour.Material != 1 || contour.Indices[0] != contour.Indices[len(contour.Indices)-1] {
				t.Fatal("packed contour lost material or closing edge", model.Name)
			}
		}
	}
	bad := append([]byte(nil), data...)
	bad[0xad6e-0x9000+4] = 255
	if _, err := source.AuthorModels(bad); err == nil {
		t.Fatal("out-of-bank vertex accepted")
	}
	sines, err := source.AuthorSines(data)
	if err != nil || len(sines) != 2048 || sines[0] != 0 || sines[1] != 100 {
		t.Fatal("original wave table changed", err)
	}
}
