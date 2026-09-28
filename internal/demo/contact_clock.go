package demo

import "github.com/olivierh59500/go-mentalhangover/internal/source"

type contactPose struct {
	x, y, frame, height int
	visible             bool
}
type contactClock struct {
	data                  source.ContactData
	angles                [3]int16
	offset                [2]int16
	head, depth           int
	phase, local          int
	ballLevel, titleLevel int
	poses                 [130]contactPose
	done                  bool
}

var contactPhaseFrames = [6]int{65, 249, 33, 749, 33, 65}

func newContactClock(data source.ContactData) *contactClock {
	return &contactClock{data: data, angles: [3]int16{0, 180, 0}, depth: 3250, local: -1}
}

func (c *contactClock) Step() bool {
	if c.done {
		return false
	}
	c.local++
	if c.local >= contactPhaseFrames[c.phase] {
		c.phase++
		c.local = 0
	}
	if c.phase == len(contactPhaseFrames) {
		c.done = true
		return false
	}
	c.ballLevel, c.titleLevel = 32, 0
	switch c.phase {
	case 0:
		c.ballLevel = 32 - (64-c.local)/2
	case 2:
		c.titleLevel = c.local
	case 3:
		c.titleLevel = 32
	case 4:
		c.titleLevel = 32 - c.local
	case 5:
		c.ballLevel = (64 - c.local) / 2
	}
	if c.phase >= 3 {
		c.angles[1] += 3
		c.angles[2] += 2
	}
	for i := range c.angles {
		if c.angles[i] > 718 {
			c.angles[i] -= 720
		} else if c.angles[i] < 0 {
			c.angles[i] += 720
		}
	}
	velocity := contactVelocity(c.angles, c.data.Sines)
	c.offset[0] = int16(uint16(c.offset[0]) + uint16(velocity[0]))
	c.offset[1] = int16(uint16(c.offset[1]) + uint16(velocity[1]))
	c.depth -= int(velocity[2])
	for c.depth < 3225 {
		c.depth += 25
		c.head--
	}
	for c.depth > 3250 {
		c.depth -= 25
		c.head++
	}
	c.head = (c.head%130 + 130) % 130
	for i := range c.poses {
		point := c.data.Points[(c.head+i)%130]
		depth := c.depth - i*25
		diameter := min(15, 8192/(depth+1)) + 1
		factor := int32(263680 / (depth + 90))
		x := int32(int16((uint16(point.X)+uint16(c.offset[0]))&1023) - 512)
		y := int32(int16((uint16(point.Y)+uint16(c.offset[1]))&1023) - 512)
		px, py := int(int16(x*factor>>9))+176, int(int16(y*factor>>9))+(16-diameter)/2+159
		c.poses[i] = contactPose{x: px, y: py - 26, frame: 16 - diameter, height: diameter,
			visible: x != 0 && y != 0 && px >= 0 && px < 336 && py >= 0 && py < 302}
	}
	return true
}

func contactVelocity(angles [3]int16, sines []int16) [3]int16 {
	var wave [6]int16
	for i, a := range angles {
		index := int(uint16(a)&0xfffe) / 2
		// The source can produce the guard offset -1 on an odd phase step.
		// Its high lookup lands in the cleared screen-buffer guard area.
		if index < len(sines) {
			wave[i*2] = sines[index]
		}
		cos := uint16(a) + 180
		if int16(cos) > 718 {
			cos -= 720
		} else if int16(cos) < 0 {
			cos += 720
		}
		wave[i*2+1] = sines[int(cos&0xfffe)/2]
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(v int32) int16 { return int16(v >> 16) }
	return [3]int16{int16(int32(high(mul(cx, cy))) >> 10),
		int16(int32(high(mul(sx, cz)-mul(high(mul(sz, sy)<<1), cx))) >> 10),
		int16(int32(high(mul(sx, sz)+mul(high(mul(cz, sy)<<1), cx))) >> 8)}
}
