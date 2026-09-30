package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

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
	plane                      *sprites.IndexedPointPlane
	velocity                   *motion.WordEulerVelocity
}

func newStarPageClock(data source.StarPages) *starPageClock {
	c := &starPageClock{data: data, angles: [3]int16{0, -180, 0}, frame: -1, page: -1}
	var err error
	c.plane, err = newStarPointPlane(data.Points, func() [3]int16 { return c.offset })
	if err != nil {
		panic(err)
	}
	c.masks = c.plane.Masks()
	c.velocity, err = motion.NewWordEulerVelocity(motion.WordEulerVelocityConfig{Sines: data.Sines, Period: 720, Quantum: 2, Quarter: 180, OutputShift: [3]uint8{11, 11, 9}, PhasePolicy: motion.WordPhaseNormalized})
	if err != nil {
		panic(err)
	}
	return c
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
	velocity, ok := c.velocity.Sample(c.angles)
	if !ok {
		panic("demo: invalid star velocity phase")
	}
	for i, v := range velocity {
		c.offset[i] = int16(uint16(c.offset[i]) + uint16(v))
	}
	if err := c.plane.Sample(); err != nil {
		panic(err)
	}
	c.touched = c.plane.Touched()
	return true
}

// starPageVelocity translates only the three matrix products consumed as
// source velocities, preserving their distinct eleven/nine-bit truncations.
func starPageVelocity(angles [3]int16, sines []int16) [3]int16 {
	velocity, err := motion.NewWordEulerVelocity(motion.WordEulerVelocityConfig{Sines: sines, Period: 720, Quantum: 2, Quarter: 180, OutputShift: [3]uint8{11, 11, 9}, PhasePolicy: motion.WordPhaseNormalized})
	if err != nil {
		panic(err)
	}
	v, ok := velocity.Sample(angles)
	if !ok {
		panic("demo: invalid star velocity phase")
	}
	return v
}

// projectStarPages keeps the original depth reciprocal lookup and two-plane
// XOR, including cancellations when particles occupy the same output pixel.
func projectStarPages(points []source.Point3, offset [3]int16, masks []byte, touched *[]int) {
	plane, err := newStarPointPlane(points, func() [3]int16 { return offset })
	if err != nil {
		panic(err)
	}
	defer plane.Close()
	if err := plane.Sample(); err != nil {
		panic(err)
	}
	copy(masks, plane.Masks())
	*touched = append((*touched)[:0], plane.Touched()...)
}
