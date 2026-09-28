package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestStarPageBanksKeepBothOriginalMessagesAndAllParticles(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/textured-cube.bin")
	pages, err := source.StarPageData(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages.Points) != 260 || len(pages.Sines) != 360 || len(pages.Cards) != 2 {
		t.Fatal("star-page assets are incomplete")
	}
	if len(pages.Cards[0].Lines) != 17 || pages.Cards[0].Lines[0] != "SPECIAL GREETINGS" ||
		len(pages.Cards[1].Lines) != 11 || pages.Cards[1].Lines[0] != "OUR GLORY MEMBERS ARE:" {
		t.Fatal("original greetings or members were truncated")
	}
	if pages.Advances[0] != 11 || pages.Advances[1] != 5 || pages.Font.Bounds().Dx() != 960 {
		t.Fatal("source small-font addresses are incorrect")
	}
}
