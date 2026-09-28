package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

type circleClock struct {
	data                            source.CircleData
	phase                           int
	local                           int
	angle                           byte
	cursor, pause                   int
	radiusPhase, profilePhase       byte
	radiusStep, factor, profileStep byte
	base                            int
	rows                            [11]int
	letters                         [8]byte
	level                           int
	active                          bool
	done                            bool
}

func newCircleClock(data source.CircleData) *circleClock {
	return &circleClock{data: data, local: -1, radiusPhase: 64, profilePhase: 192, radiusStep: 2, factor: 22, profileStep: 4}
}

func (c *circleClock) Step() bool {
	if c.done {
		return false
	}
	c.local++
	if c.phase == 0 {
		c.level = c.local
		if c.local == 32 {
			c.phase = 1
			c.local = -1
		}
		return true
	}
	if c.phase == 2 {
		c.level = 32 - c.local
		if c.local == 32 {
			c.done = true
		}
		return true
	}
	c.active = true
	c.radiusPhase += c.radiusStep
	c.profilePhase += c.profileStep
	c.base = ((int(c.data.Profile[c.radiusPhase]) + 127) * int(c.factor) & 0xff00) / 256
	for i := range c.rows {
		// The source offset retains its 0xff high byte and samples the preceding
		// coordinate wave, including after the low phase byte wraps.
		c.rows[i] = ((20 - i*4) * int(c.data.Wave[c.profilePhase]) >> 7) + 20
	}
	if c.pause > 0 {
		c.pause--
	} else {
		c.angle -= 2
		if int8(c.angle) >= 0 {
			c.angle += 30
			c.cursor++
			if c.cursor >= len(c.data.Text) {
				c.finish()
				return true
			}
			if c.data.Text[c.cursor] == '|' {
				c.pause = (int(c.data.Text[c.cursor+1]) - 32) * 25
				c.cursor += 2
			}
			if c.data.Text[c.cursor] == '>' {
				v := c.data.Text[c.cursor+1 : c.cursor+6]
				c.radiusStep, c.factor, c.profileStep, c.radiusPhase, c.profilePhase = v[0], v[1], v[2], v[3], v[4]
				c.cursor += 6
			}
		}
	}
	index := c.cursor
	for slot := range c.letters {
		for {
			if index >= len(c.data.Text) || c.data.Text[index] >= 128 {
				c.finish()
				return true
			}
			value := c.data.Text[index]
			index++
			if value == '|' {
				index++
				continue
			}
			if value == '>' {
				index += 5
				continue
			}
			c.letters[slot] = value
			break
		}
	}
	return true
}

func (c *circleClock) finish() { c.phase = 2; c.local = -1; c.active = false; c.level = 32 }

func (c *circleClock) Point(slot int, p source.PolarGlyphPoint) source.Point2 {
	angle := byte(int(c.angle) + 100 + (slot+1)*30 + int(p.Angle))
	radius := (c.base + c.rows[p.Row]) * 2
	return source.Point2{X: int16(int(c.data.Wave[byte(0-angle)])*radius>>7) + 173,
		Y: int16(int(c.data.Wave[byte(64-angle)])*radius>>7) + 108}
}
