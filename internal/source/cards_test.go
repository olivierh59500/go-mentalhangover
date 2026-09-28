package source_test

import (
	"reflect"
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestOriginalCardsRetainBlankLinesAndVerticalOrigins(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/authors-ribbon.bin")
	cards, err := source.AuthorCards(data)
	if err != nil || len(cards) != 7 {
		t.Fatal("original text cards missing", err)
	}
	if cards[0].Y != 120 || !reflect.DeepEqual(cards[0].Lines, []string{"PRESENTS"}) ||
		cards[1].Y != 108 || !reflect.DeepEqual(cards[1].Lines, []string{"A NEW", "DEMONSTRATION!"}) ||
		cards[6].Y != 85 || !reflect.DeepEqual(cards[6].Lines, []string{"AND NOW...", "", "FOLLOW", "THE SIGN!!"}) {
		t.Fatal("authored card layout changed", cards)
	}
}

func TestResidentInterludesRetainAuthoredLines(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/resident.bin")
	cards, err := source.ResidentCards(data)
	if err != nil || len(cards) != 4 {
		t.Fatal("resident interludes missing", err)
	}
	if cards[0].Y != 60 || !reflect.DeepEqual(cards[0].Lines,
		[]string{"THIS ISN'T", "A FUCKING", "MEGADEMO", "", "THIS IS A", "SCOOPEX DEMO"}) {
		t.Fatal("original interlude spacing changed", cards[0])
	}
}

func TestCopperPaletteUsesNibbleTruncationAndBoundarySamples(t *testing.T) {
	word := uint16(0x4ad)
	if source.PaletteWord(word, source.PaletteFromWhite, 0) != 0xfff ||
		source.PaletteWord(word, source.PaletteFromWhite, 32) != word ||
		source.PaletteWord(word, source.PaletteToBlack, 8) != 0x256 ||
		source.PaletteWord(word, source.PaletteToBlack, 0) != 0 {
		t.Fatal("integer palette transition changed")
	}
	cases := []struct {
		tick  int
		mode  source.PaletteMode
		value int
	}{
		{0, source.PaletteGray, 0}, {15, source.PaletteGray, 15},
		{16, source.PaletteFromWhite, 0}, {48, source.PaletteFromWhite, 32},
		{49, source.PaletteTarget, 0}, {149, source.PaletteTarget, 0},
		{150, source.PaletteToBlack, 15}, {165, source.PaletteToBlack, 0},
	}
	for _, c := range cases {
		mode, value := source.CardPalette(c.tick, 101)
		if mode != c.mode || value != c.value {
			t.Fatal("card transition boundary changed", c.tick, mode, value)
		}
	}
}
