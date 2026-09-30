package demo

import (
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

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
	program                                                *scrolltext.InsertionProgram
	done                                                   bool
}

func newBOBClock(data source.BOBData) *bobClock {
	c := &bobClock{data: data, speed: 4, targetSpeed: 4, depth: 4100, target: 4100}
	tokens := make([]scrolltext.InsertionToken, len(data.Tokens))
	for i, t := range data.Tokens {
		tokens[i] = scrolltext.InsertionToken{Command: t.Kind, Payload: t.Values[:]}
		if t.Kind == 't' {
			tokens[i].Glyph = true
			tokens[i].Rune = t.Letter
			tokens[i].Advance = data.Advances[int(t.Letter)-32]
		}
	}
	var err error
	c.program, err = scrolltext.NewInsertionProgram(scrolltext.InsertionProgramConfig{Tokens: tokens, Speed: 4, TargetSpeed: 4, SpeedStep: 1, Divisor: 2, Entry: 368, AlignBias: 14, AlignMask: 15, RetireBefore: -16, OnCommand: c.command})
	if err != nil {
		panic(err)
	}
	c.origins = c.program.State().Origins
	return c
}
func (c *bobClock) command(p *scrolltext.InsertionProgram, t scrolltext.InsertionToken) error {
	switch t.Command {
	case '>':
		return p.SetTargetSpeed(int(t.Payload[0]) * 2)
	case '|':
		return p.SetPause(int(t.Payload[0]) * 50)
	case 'a':
		c.target = 1100
	case 'd':
		c.target = 4100
	case 'b':
		v := t.Payload
		c.count, c.model = int(v[0]), int(v[13])
		c.stepX, c.stepY = int16(int8(v[1]))*8, int16(int8(v[2]))*4
		c.gapX, c.gapY = int16(int8(v[3]))*8, int16(int8(v[4]))*4
		c.phaseX, c.phaseY = int16(v[5]>>3), int16(v[6])*4
		for i := 0; i < 3; i++ {
			c.rotation[i] = int16(int8(v[i+7])) * 2
			c.angles[i] = int16(int8(v[i+10])) * 2
		}
	}
	return nil
}

func (c *bobClock) Step() bool {
	if c.done {
		return false
	}
	active, err := c.program.Step()
	if err != nil {
		panic(err)
	}
	s := c.program.State()
	c.cursor, c.remaining, c.speed, c.targetSpeed, c.pause, c.distance = s.Cursor, s.Remaining, s.Speed, s.TargetSpeed, s.Pause, s.Distance
	c.first, c.fetched = s.First, s.Fetched
	c.done = s.Finished
	if !active {
		return false
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
