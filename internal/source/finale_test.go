package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestFinaleBanksDecodeAllOriginalMaterialsAndColumnRows(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/checkerboard.bin")
	finale, err := source.FinalePart(data)
	if err != nil {
		t.Fatal(err)
	}
	if finale.Logo.Bounds().Dx() != 320 || finale.Logo.Bounds().Dy() != 87 || len(finale.Columns) != 185*48 ||
		len(finale.Balls) != 6 || len(finale.Colors) != 90 || len(finale.Particles) != 24 || len(finale.Bounce) != 256 {
		t.Fatal("finale assets are incomplete")
	}
	for _, ball := range finale.Balls {
		if ball.Bounds().Dx() != 960 {
			t.Fatal("sphere size row was truncated")
		}
	}
}
