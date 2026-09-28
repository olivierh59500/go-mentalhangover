package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

// bobClock retains the source's look-ahead insertion and binary controls. The
// text, copies, rotation and camera have independent clocks during text pauses.
type bobClock struct {
	data                                                   source.BOBData
	cursor, remaining, speed, targetSpeed, pause, distance int
	origins                                                []int
	first, fetched                                         int
	count, model                                           int
	phaseX, phaseY, stepX, stepY, gapX, gapY               int16
	angles, rotation                                       [3]int16
	depth, target                                          int16
	done                                                   bool
}

func newBOBClock(data source.BOBData) *bobClock {
	return &bobClock{data: data, speed: 4, targetSpeed: 4, depth: 4100, target: 4100, origins: make([]int, len([]rune(data.Text)))}
}

func (c *bobClock) Step() bool {
	if c.done {
		return false
	}
	if c.pause > 0 {
		c.pause--
	} else {
		c.remaining -= c.speed / 2
		if c.remaining < 0 {
			inserted := false
			for c.cursor < len(c.data.Tokens) && !inserted {
				token := c.data.Tokens[c.cursor]
				c.cursor++
				switch token.Kind {
				case '>':
					c.targetSpeed = int(token.Values[0]) * 2
				case '|':
					c.pause = int(token.Values[0]) * 50
				case 'a':
					c.target = 1100
				case 'd':
					c.target = 4100
				case 'b':
					v := token.Values
					c.count, c.model = int(v[0]), int(v[13])
					c.stepX, c.stepY = int16(int8(v[1]))*8, int16(int8(v[2]))*4
					c.gapX, c.gapY = int16(int8(v[3]))*8, int16(int8(v[4]))*4
					c.phaseX, c.phaseY = int16(v[5]>>3), int16(v[6])*4
					for i := 0; i < 3; i++ {
						c.rotation[i] = int16(int8(v[i+7])) * 2
						c.angles[i] = int16(int8(v[i+10])) * 2
					}
				case 't':
					// Only the first three masked words of the source glyph blit
					// participate; the insertion is ahead of the visible viewport.
					c.origins[c.fetched] = c.distance + 368 + ((c.remaining + 14) & 15)
					c.fetched++
					c.remaining += c.data.Advances[int(token.Letter)-32]
					inserted = true
				}
			}
			if !inserted {
				c.done = true
				return false
			}
		}
		c.distance += c.speed / 2
		if c.speed < c.targetSpeed {
			c.speed++
		} else if c.speed > c.targetSpeed {
			c.speed--
		}
	}
	for c.first < c.fetched && c.origins[c.first]-c.distance < -16 {
		c.first++
	}
	for i, delta := range c.rotation {
		c.angles[i] = int16(uint16(c.angles[i]) + uint16(delta))
	}
	if c.depth > c.target {
		c.depth -= 200
	} else if c.depth != c.target && c.depth != 4100 {
		c.depth += 200
	}
	c.phaseX = int16(uint16(c.phaseX) + uint16(c.stepX))
	c.phaseY = int16(uint16(c.phaseY) + uint16(c.stepY))
	return true
}

func (c *bobClock) Offset(index int) (int, int) {
	x := int16(uint16(c.phaseX) + uint16(c.gapX)*uint16(index+1))
	y := int16(uint16(c.phaseY) + uint16(c.gapY)*uint16(index+1))
	return c.data.X[(uint16(x)&0xff8)/8] - 16, c.data.Y[(uint16(y)&0x7fc)/4] + 34
}
