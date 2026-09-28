package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

type perspectiveClock struct {
	data          source.PerspectiveData
	phase, local  int
	position      byte
	cursor        int
	letters       [8]byte
	offset        [3]int16
	level, border int
	done          bool
}

func newPerspectiveClock(data source.PerspectiveData) *perspectiveClock {
	return &perspectiveClock{data: data, local: -1}
}

func (c *perspectiveClock) Step() bool {
	if c.done {
		return false
	}
	c.local++
	if c.phase == 0 {
		c.border = (c.local/7 + 1) * 0x111
		if c.local == 55 {
			c.phase = 1
			c.local = -1
		}
		return true
	}
	if c.phase == 1 {
		c.level = c.local + 1
		if c.local == 31 {
			c.phase = 2
			c.local = -1
		}
		c.advancePoints()
		return true
	}
	if c.phase == 3 {
		c.level = 31 - c.local
		c.advancePoints()
		if c.local == 31 {
			c.phase = 4
			c.local = -1
		}
		return true
	}
	if c.phase == 4 {
		c.border = (7 - c.local/7) * 0x111
		if c.local == 48 {
			c.done = true
		}
		return true
	}
	c.position -= 2
	if int8(c.position) < 0 {
		c.position += 22
		c.cursor++
	}
	for i := range c.letters {
		if c.cursor+i >= len(c.data.Text) || c.data.Text[c.cursor+i] >= 128 {
			c.phase = 3
			c.local = -1
			return true
		}
		c.letters[i] = c.data.Text[c.cursor+i]
	}
	c.advancePoints()
	return true
}

func (c *perspectiveClock) advancePoints() {
	c.offset[0] = int16(uint16(c.offset[0]) - 6)
	c.offset[1] = int16(uint16(c.offset[1]) + 3)
	c.offset[2] = int16(uint16(c.offset[2]) - 10)
}

// Perspective outlines use the source's precomputed eleven-row, 280-column
// lookup. Integer division and each row's start values remain independent.
func (c *perspectiveClock) Point(slot int, p source.PolarGlyphPoint) source.Point2 {
	column := int(c.position) + slot*22 + int(p.Angle)
	row := int(p.Row)
	denominator := 330 + row*10 + column*5
	x := -81920 - row*4096 + column*1536
	y := 71680 - row*2560 - column*819
	return source.Point2{X: int16(x/denominator) + 173, Y: int16(y/denominator) + 108}
}
