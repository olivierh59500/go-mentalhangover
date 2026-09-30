package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type finalBall struct {
	x, y, size, material int
	visible              bool
}
type finaleClock struct {
	data                           source.FinaleData
	tick, floorTick, paletteOffset int
	ballCount, depth               int
	particles                      [24][4]byte
	poses                          [24]finalBall
	logoLevel, reveal              int
	queue                          *motion.RecycledQueue[[4]byte, finalBall]
}

func newFinaleClock(data source.FinaleData) *finaleClock {
	c := &finaleClock{data: data, tick: -1, ballCount: 1, reveal: 1}
	items := make([][4]byte, len(c.particles))
	copy(items, data.Particles)
	var err error
	c.queue, err = motion.NewRecycledQueue(motion.RecycledQueueConfig[[4]byte, finalBall]{Items: items, Count: 1, Depth: 0, Step: 28, Spacing: 100, DepthWrap: 2400, Grow: true,
		Recycle: func(tick int, p *[4]byte) { p[0] = 0; p[2] = byte(tick*73 + 19) }, Project: c.projectBall,
	})
	if err != nil {
		panic(err)
	}
	copy(c.particles[:], c.queue.Items())
	return c
}
func (c *finaleClock) projectBall(_ int, depth int, p *[4]byte) (finalBall, error) {
	phase := p[0]
	p[0] += byte(p[1]&3) + 5
	denominator := 2560 - depth
	size := min(32, 10240/(denominator-100))
	y := (140 - int(c.data.Bounce[phase])) * 512 / denominator
	x := int(int8(p[2]))*1024/denominator + 192
	return finalBall{x: x - 32, y: y - size/2 + 138, size: size, material: int(p[3]), visible: y < 155 && x < 368 && x > 16 && size >= 3}, nil
}

func (c *finaleClock) Step() {
	c.tick++
	c.logoLevel = min(64, c.tick+1)
	if c.tick < 64 {
		return
	}
	c.floorTick++
	if c.floorTick&15 == 0 {
		c.paletteOffset -= 34
		if c.paletteOffset <= 0 {
			c.paletteOffset += 1530
		}
		if c.tick < 365 {
			c.reveal = min(17, c.reveal+2)
		}
	}
	if c.tick < 365 {
		return
	}
	if err := c.queue.Step(c.tick); err != nil {
		panic(err)
	}
	c.depth, c.ballCount = c.queue.Depth(), c.queue.Count()
	copy(c.particles[:], c.queue.Items())
	copy(c.poses[:], c.queue.Poses())
}

func (c *finaleClock) rowColors(y int) (uint16, uint16) {
	phase := c.floorTick & 15
	wrap := func(index int) int {
		if index > 1530 {
			index -= 1530
		}
		return index
	}
	a, b := c.paletteOffset, wrap(c.paletteOffset+765)
	first, second := b, a
	for row := 0; row <= 16; row++ {
		line := 172 + 71680/(2560-phase*16-row*128) - 34
		if y < line {
			break
		}
		if row >= c.reveal && c.tick < 365 {
			return 0, 0
		}
		first, second = a, b
		a = wrap(a + 18)
		b = wrap(b + 18)
		a, b = b, a
	}
	scale := func(index int) uint16 {
		if index == 1530 {
			// The source permits the endpoint, which aliases the first two
			// bytes (4, 4) of its adjacent sphere-size table. Keep that color
			// explicitly without reading outside the authored palette bank.
			return 0x404
		}
		word := c.data.Colors[index/17]
		level := index%17 + 5
		var result uint16
		for shift := 0; shift <= 8; shift += 4 {
			result |= uint16((int(word>>shift&15)*level)/22) << shift
		}
		return result
	}
	return scale(first), scale(second)
}
