package demo

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestFinaleRunsCompleteColorAndParticleCyclesWithoutEscapingBanks(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/checkerboard.bin")
	data, err := source.FinalePart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newFinaleClock(data)
	for frame := 0; frame < 8000; frame++ {
		c.Step()
		for y := 166; y < Height; y++ {
			c.rowColors(y)
		}
		if c.ballCount > 24 {
			t.Fatal("ball depth queue escaped")
		}
		for _, p := range c.poses[:c.ballCount] {
			if p.visible && (p.size < 3 || p.size > 32 || p.material < 0 || p.material > 5) {
				t.Fatal("ball atlas lookup escaped", p)
			}
		}
	}
}
