package demo

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// Only the source's beam-time random seed is substituted in the CPU fixture.
// Queue state, divisions, bounce samples and copper colors execute unchanged.
func TestFinaleQueueProjectionAndCopperColorsMatchOriginalCPU(t *testing.T) {
	bytes, _ := assets.Files.ReadFile("raw/checkerboard.bin")
	data, err := source.FinalePart(bytes)
	if err != nil {
		t.Fatal(err)
	}
	c := newFinaleClock(data)
	poses := originalLateRows(t, "finale-poses")
	colors := originalLateRows(t, "finale-rows")
	poseIndex, colorIndex := 0, 0
	states := originalLateRows(t, "finale-clock")
	if len(states) != 1984 {
		t.Fatal("incomplete source ending fixture")
	}
	for _, row := range states {
		tick := lateInt(t, row[0])
		for c.tick < tick {
			c.Step()
		}
		got := []int{tick, c.floorTick, c.paletteOffset, c.ballCount, c.depth}
		for i, value := range got {
			if value != lateInt(t, row[i]) {
				t.Fatalf("ending tick %d field %d: Go %d, CPU %s", tick, i, value, row[i])
			}
		}
		for poseIndex < len(poses) && lateInt(t, poses[poseIndex][0]) == tick {
			row := poses[poseIndex]
			p := c.poses[lateInt(t, row[1])]
			x, y, size := lateInt(t, row[2]), lateInt(t, row[3]), lateInt(t, row[4])
			visible := y < 155 && x < 368 && x > 16 && size >= 3
			if p.x+32 != x || p.y-138+size/2 != y || p.size != size || p.material != lateInt(t, row[5]) || p.visible != visible {
				t.Fatalf("ending tick %d sphere %s: Go %+v, CPU %v", tick, row[1], p, row)
			}
			poseIndex++
		}
		for colorIndex < len(colors) && lateInt(t, colors[colorIndex][0]) == tick {
			row := colors[colorIndex]
			y := lateInt(t, row[1])
			a, b := c.rowColors(y)
			if int(a) != lateInt(t, row[2]) || int(b) != lateInt(t, row[3]) {
				t.Fatalf("ending tick %d row %d: Go %03x,%03x; CPU %v", tick, y, a, b, row)
			}
			colorIndex++
		}
	}
	if poseIndex != len(poses) || colorIndex != len(colors) {
		t.Fatal("ending fixture samples left unchecked")
	}
}

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
