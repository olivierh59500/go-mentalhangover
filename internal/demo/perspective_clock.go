package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type perspectiveClock struct {
	data          source.PerspectiveData
	phase, local  int
	position      byte
	cursor        int
	letters       [8]byte
	offset        [3]int16
	level, border int
	done          bool
	pointMasks    [Width * 200]byte
	touched       []int
	window        *scrolltext.ByteWindow
	projection    *motion.RationalGrid
}

func newPerspectiveClock(data source.PerspectiveData) *perspectiveClock {
	c := &perspectiveClock{data: data, local: -1, touched: make([]int, 0, len(data.Points))}
	var err error
	c.window, err = scrolltext.NewByteWindow(scrolltext.ByteWindowConfig{Text: data.Text, Slots: len(c.letters), Step: -2, Advance: 22, Crossing: scrolltext.BelowZero})
	if err != nil {
		panic(err)
	}
	c.projection, err = motion.NewRationalGrid(motion.RationalGridConfig{X: [3]int64{-81920, -4096, 1536}, Y: [3]int64{71680, -2560, -819}, Denominator: [3]int64{330, 10, 5}, Center: [2]int64{173, 108}, WordBits: 16})
	if err != nil {
		panic(err)
	}
	c.rasterizePoints()
	return c
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
	if _, err := c.window.Step(); err != nil {
		panic(err)
	}
	c.position, c.cursor = c.window.Position(), c.window.Cursor()
	copy(c.letters[:], c.window.Letters())
	if c.window.Finished() {
		c.phase = 3
		c.local = -1
		return true
	}
	c.advancePoints()
	return true
}

func (c *perspectiveClock) advancePoints() {
	c.offset[0] = int16(uint16(c.offset[0]) - 6)
	c.offset[1] = int16(uint16(c.offset[1]) + 3)
	c.offset[2] = int16(uint16(c.offset[2]) - 10)
	c.rasterizePoints()
}

// The original two point planes combine coincident stars with bitwise OR.
// Clear only touched pixels and reuse the bounded index list on each update.
func (c *perspectiveClock) rasterizePoints() {
	for _, index := range c.touched {
		c.pointMasks[index] = 0
	}
	c.touched = c.touched[:0]
	for _, p := range c.data.Points {
		z := (uint16(p.Z) + uint16(c.offset[2])) & 2047
		factor := int32(263680 / (int(z) + 390))
		x := int32(int16((uint16(p.X)+uint16(c.offset[0]))&1023) - 512)
		y := int32(int16((uint16(p.Y)+uint16(c.offset[1]))&511) - 256)
		px, py := int(int16(x*factor>>9))+176, int(int16(y*factor>>9))+100
		if px < 0 || px >= Width || py < 0 || py >= 199 {
			continue
		}
		mask := byte(2)
		if z < 1000 {
			mask = 1
		} else if z >= 1700 {
			mask = 3
		}
		index := py*Width + px
		if c.pointMasks[index] == 0 {
			c.touched = append(c.touched, index)
		}
		c.pointMasks[index] |= mask
	}
}

// Perspective outlines use the source's precomputed eleven-row, 280-column
// lookup. Integer division and each row's start values remain independent.
func (c *perspectiveClock) Point(slot int, p source.PolarGlyphPoint) source.Point2 {
	point, ok := c.projection.Point(int64(c.position)+int64(slot)*22+int64(p.Angle), int64(p.Row))
	if !ok {
		panic("demo: invalid perspective glyph projection")
	}
	return source.Point2{X: int16(point.X), Y: int16(point.Y)}
}
