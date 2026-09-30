package demo

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/democonstructionkit/sprites"
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
	pointMasks    []byte
	touched       []int
	window        *scrolltext.ByteWindow
	projection    *motion.RationalGrid
	plane         *sprites.IndexedPointPlane
}

func newPerspectiveClock(data source.PerspectiveData, white ...*ebiten.Image) *perspectiveClock {
	c := &perspectiveClock{data: data, local: -1, touched: make([]int, 0, len(data.Points))}
	var err error
	c.window, err = scrolltext.NewByteWindow(scrolltext.ByteWindowConfig{Text: data.Text, Slots: len(c.letters), Step: -2, Advance: 22, Crossing: scrolltext.BelowZero})
	if err != nil {
		panic(err)
	}
	var pixel *ebiten.Image
	if len(white) > 0 {
		pixel = white[0]
	}
	c.plane, err = newPerspectivePointPlane(data.Points, func() [3]int16 { return c.offset }, pixel)
	if err != nil {
		panic(err)
	}
	c.pointMasks = c.plane.Masks()
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
	if err := c.plane.Sample(); err != nil {
		panic(err)
	}
	c.touched = c.plane.Touched()
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
