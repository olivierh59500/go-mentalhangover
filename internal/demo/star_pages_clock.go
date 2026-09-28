package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

type starPagePhase struct {
	name                string
	frames, page        int
	stars, text         int
	fadeStars, fadeText bool
	down                bool
}

var starPagePhases = []starPagePhase{
	{"stars-in", 32, -1, 32, 0, true, false, false},
	{"stars-first-hold", 179, -1, 32, 0, false, false, false},
	{"greetings-in", 32, 0, 32, 31, false, true, false},
	{"greetings-hold", 539, 0, 32, 31, false, false, false},
	{"greetings-out", 32, 0, 32, 31, false, true, true},
	{"stars-second-hold", 179, -1, 32, 0, false, false, false},
	{"members-in", 32, 1, 32, 31, false, true, false},
	{"members-hold", 539, 1, 32, 31, false, false, false},
	{"members-out", 32, 1, 32, 31, false, true, true},
	{"stars-last-hold", 119, -1, 32, 0, false, false, false},
	{"stars-out", 32, -1, 31, 0, true, false, true},
}

type starPageClock struct {
	data                       source.StarPages
	angles                     [3]int16
	offset                     [3]int16
	masks                      []byte
	touched                    []int
	frame, phase, local        int
	page, starLevel, textLevel int
	done                       bool
}

func newStarPageClock(data source.StarPages) *starPageClock {
	return &starPageClock{data: data, angles: [3]int16{0, -180, 0}, masks: make([]byte, Width*286),
		touched: make([]int, 0, 260), frame: -1, page: -1}
}

func (c *starPageClock) Step() bool {
	if c.done {
		return false
	}
	c.frame++
	c.local++
	if c.frame == 0 {
		c.local = 0
	}
	if c.local >= starPagePhases[c.phase].frames {
		c.phase++
		c.local = 0
	}
	if c.phase >= len(starPagePhases) {
		c.done = true
		return false
	}
	p := starPagePhases[c.phase]
	c.page, c.starLevel, c.textLevel = p.page, p.stars, p.text
	if p.fadeStars {
		c.starLevel = c.local + 1
		if p.down {
			c.starLevel = 31 - c.local
		}
	}
	if p.fadeText {
		c.textLevel = c.local
		if p.down {
			c.textLevel = 31 - c.local
		}
	}
	if c.phase >= 2 {
		c.angles[2] += 2
	}
	if c.phase >= 4 {
		c.angles[1] += 4
	}
	for i := range c.angles {
		if c.angles[i] > 718 {
			c.angles[i] -= 720
		} else if c.angles[i] < 0 {
			c.angles[i] += 720
		}
	}
	velocity := starPageVelocity(c.angles, c.data.Sines)
	for i, v := range velocity {
		c.offset[i] = int16(uint16(c.offset[i]) + uint16(v))
	}
	for _, index := range c.touched {
		c.masks[index] = 0
	}
	c.touched = c.touched[:0]
	projectStarPages(c.data.Points, c.offset, c.masks, &c.touched)
	return true
}

// starPageVelocity translates only the three matrix products consumed as
// source velocities, preserving their distinct eleven/nine-bit truncations.
func starPageVelocity(angles [3]int16, sines []int16) [3]int16 {
	var wave [6]int16
	for i, a := range angles {
		n := int(a) &^ 1
		wave[i*2] = sines[n/2]
		wave[i*2+1] = sines[((n+180)%720)/2]
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(v int32) int16 { return int16(v >> 16) }
	return [3]int16{
		int16(int32(high(mul(cx, cy))) >> 11),
		int16(int32(high(mul(sx, cz)-mul(high(mul(sz, sy)<<1), cx))) >> 11),
		int16(int32(high(mul(sx, sz)+mul(high(mul(cz, sy)<<1), cx))) >> 9),
	}
}

// projectStarPages keeps the original depth reciprocal lookup and two-plane
// XOR, including cancellations when particles occupy the same output pixel.
func projectStarPages(points []source.Point3, offset [3]int16, masks []byte, touched *[]int) {
	for _, p := range points {
		x := int32(int16((uint16(p.X)+uint16(offset[0]))&511) - 256)
		y := int32(int16((uint16(p.Y)+uint16(offset[1]))&511) - 256)
		z := (uint16(p.Z) + uint16(offset[2])) & 2047
		factor := int32(262144 / (int(z) + 10))
		px, py := int(int16(x*factor>>9))+176, int(int16(y*factor>>9))+143
		if px < 0 || px >= Width || py < 0 || py >= 286 {
			continue
		}
		mask := byte(0)
		if z < 1700 {
			mask |= 1
		}
		if z > 1000 {
			mask |= 2
		}
		index := py*Width + px
		if masks[index] == 0 {
			*touched = append(*touched, index)
		}
		masks[index] ^= mask
	}
}
