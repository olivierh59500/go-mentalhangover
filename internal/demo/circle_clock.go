package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

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
	window                          *scrolltext.ByteWindow
	projection                      *motion.TablePolar
}

func newCircleClock(data source.CircleData) *circleClock {
	c := &circleClock{data: data, local: -1, radiusPhase: 64, profilePhase: 192, radiusStep: 2, factor: 22, profileStep: 4}
	var err error
	c.window, err = scrolltext.NewByteWindow(scrolltext.ByteWindowConfig{Text: data.Text, Slots: len(c.letters), Step: -2, Advance: 30, Crossing: scrolltext.NonNegative,
		Commands: map[byte]scrolltext.ByteCommand{
			'|': {Payload: 1, Apply: func(v []byte) (int, error) { return (int(v[0]) - 32) * 25, nil }},
			'>': {Payload: 5, Apply: func(v []byte) (int, error) {
				c.radiusStep, c.factor, c.profileStep, c.radiusPhase, c.profilePhase = v[0], v[1], v[2], v[3], v[4]
				return 0, nil
			}},
		}})
	if err != nil {
		panic(err)
	}
	wave := make([]int16, len(data.Wave))
	for i, v := range data.Wave {
		wave[i] = int16(v)
	}
	c.projection, err = motion.NewTablePolar(motion.TablePolarConfig{Wave: wave, Shift: 7, Center: [2]int64{173, 108}, Phases: [2]int{0, 64}, WordBits: 16})
	if err != nil {
		panic(err)
	}
	return c
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
	if _, err := c.window.Step(); err != nil {
		panic(err)
	}
	c.angle, c.cursor, c.pause = c.window.Position(), c.window.Cursor(), c.window.Pause()
	copy(c.letters[:], c.window.Letters())
	if c.window.Finished() {
		c.finish()
	}
	return true
}

func (c *circleClock) finish() { c.phase = 2; c.local = -1; c.active = false; c.level = 32 }

func (c *circleClock) Point(slot int, p source.PolarGlyphPoint) source.Point2 {
	angle := int64(c.angle) + 100 + int64(slot+1)*30 + int64(p.Angle)
	point, ok := c.projection.Point(angle, int64(c.base+c.rows[p.Row])*2)
	if !ok {
		panic("demo: invalid circular glyph projection")
	}
	return source.Point2{X: int16(point.X), Y: int16(point.Y)}
}
