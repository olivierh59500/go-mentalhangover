package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

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
}

func newFinaleClock(data source.FinaleData) *finaleClock {
	c := &finaleClock{data: data, tick: -1, ballCount: 1, reveal: 1}
	copy(c.particles[:], data.Particles)
	return c
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
	c.depth += 28
	if c.depth > 100 {
		c.depth -= 100
		c.ballCount = min(24, c.ballCount+1)
		last := c.particles[23]
		copy(c.particles[1:], c.particles[:23])
		c.particles[0] = last
		// The original mixes horizontal beam time into the new ball's X seed.
		// A fixed replay seed keeps captures and native playback reproducible.
		seed := byte(c.tick*73 + 19)
		c.particles[0][0] = 0
		c.particles[0][2] = seed
	}
	depth := c.depth
	for i := 0; i < c.ballCount; i++ {
		if depth > 2400 {
			depth -= 2400
		}
		p := &c.particles[i]
		p[0] += byte(p[1]&3) + 5
		denominator := 2560 - depth
		size := min(32, 10240/denominator)
		y := (140 - int(c.data.Bounce[p[0]])) * 512 / denominator
		x := int(int8(p[2]))*1024/denominator + 192
		c.poses[i] = finalBall{x: x - 32, y: y - size/2 + 138, size: size, material: int(p[3]), visible: y < 155 && x < 368 && x > 16 && size >= 3}
		depth += 100
	}
}

func (c *finaleClock) rowColors(y int) (uint16, uint16) {
	phase := c.floorTick & 15
	a, b := c.paletteOffset, (c.paletteOffset+765)%1530
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
		a = (a + 18) % 1530
		b = (b + 18) % 1530
		a, b = b, a
	}
	scale := func(index int) uint16 {
		index = (index%1530 + 1530) % 1530
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
