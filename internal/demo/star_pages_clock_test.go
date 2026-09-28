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

func starPagesForTest(t *testing.T) source.StarPages {
	t.Helper()
	data, _ := assets.Files.ReadFile("raw/textured-cube.bin")
	pages, err := source.StarPageData(data)
	if err != nil {
		t.Fatal(err)
	}
	return pages
}

func TestSteeredStarPlanesMatchOriginal68000Pixels(t *testing.T) {
	data := starPagesForTest(t)
	f, err := os.Open("testdata/original-star-pages.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		var angles, offset [3]int16
		for i := 0; i < 6; i++ {
			v, err := strconv.ParseInt(row[i+1], 10, 16)
			if err != nil {
				t.Fatal(err)
			}
			if i < 3 {
				angles[i] = int16(v)
				if angles[i] > 718 {
					angles[i] -= 720
				} else if angles[i] < 0 {
					angles[i] += 720
				}
			} else {
				offset[i-3] = int16(v)
			}
		}
		velocity := starPageVelocity(angles, data.Sines)
		for i, v := range velocity {
			offset[i] = int16(uint16(offset[i]) + uint16(v))
			want, err := strconv.ParseInt(row[i+7], 10, 16)
			if err != nil || offset[i] != int16(want) {
				t.Fatalf("star case %s velocity/offset %d differs from CPU", row[0], i)
			}
		}
		masks := make([]byte, Width*286)
		touched := make([]int, 0, 260)
		projectStarPages(data.Points, offset, masks, &touched)
		planes := make([]byte, 88*286)
		for index, mask := range masks {
			x, y := index%Width, index/Width
			if mask&1 != 0 {
				planes[y*88+x/8] |= 1 << uint(7-x%8)
			}
			if mask&2 != 0 {
				planes[y*88+44+x/8] |= 1 << uint(7-x%8)
			}
		}
		if hash := fmt.Sprintf("%x", sha256.Sum256(planes)); hash != row[10] {
			t.Fatalf("star case %s bitplanes differ from original: %s != %s", row[0], hash, row[10])
		}
	}
}

func TestStarPageProgramKeepsOriginalFadeAndHoldCountsWithoutAllocations(t *testing.T) {
	data := starPagesForTest(t)
	c := newStarPageClock(data)
	count := 0
	for c.Step() {
		if len(c.touched) > 260 {
			t.Fatal("particle raster budget escaped")
		}
		if c.frame == 211 && (c.page != 0 || c.textLevel != 0) || c.frame == 993 && (c.page != 1 || c.textLevel != 0) {
			t.Fatal("original page entry boundary changed", c.frame, c.page, c.textLevel)
		}
		count++
	}
	if count != 1747 {
		t.Fatal("original greeting operation count changed", count)
	}
	c = newStarPageClock(data)
	if got := testing.AllocsPerRun(100, func() { c.Step() }); got != 0 {
		t.Fatal("steered points allocate per update", got)
	}
}
