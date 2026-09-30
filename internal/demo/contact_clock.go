package demo

import (
	"image"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

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
	velocity              *motion.WordEulerVelocity
	queue                 *sprites.DepthQueue
	projection            *motion.WrappedPointProjection
}

var contactPhaseFrames = [6]int{65, 249, 33, 749, 33, 65}

func newContactClock(data source.ContactData) *contactClock {
	c := &contactClock{data: data, angles: [3]int16{0, 180, 0}, depth: 3250, local: -1}
	var err error
	c.velocity, err = motion.NewWordEulerVelocity(motion.WordEulerVelocityConfig{Sines: data.Sines, Period: 720, Quantum: 2, Quarter: 180, OutputShift: [3]uint8{10, 10, 8}, PhasePolicy: motion.WordPhaseGuarded})
	if err != nil {
		panic(err)
	}
	c.projection, err = motion.NewWrappedPointProjection(motion.WrappedPointProjectionConfig{
		Mask: [3]uint16{1023, 1023, 65535}, Bias: [2]int16{-512, -512},
		Numerator: 263680, DepthBias: 90, Shift: 9, Center: image.Pt(176, 159),
	})
	if err != nil {
		panic(err)
	}
	points := make([]geometry.Vec2, len(data.Points))
	for i, point := range data.Points {
		points[i] = geometry.Vec2{X: float64(point.X), Y: float64(point.Y)}
	}
	c.queue, err = sprites.NewDepthQueue(sprites.DepthQueueConfig{Points: points, Depth: 3250,
		Near: 3225, Far: 3250, Spacing: 25, UpperPolicy: sprites.QueueUpperInclusiveFirst,
		Project: c.projectBall,
	})
	if err != nil {
		panic(err)
	}
	return c
}

// projectBall binds the sixteen authored sphere sizes and their original
// visibility rules to the queue's reusable word projection and draw population.
func (c *contactClock) projectBall(slot int, point geometry.Vec2, depth int) (sprites.FieldSample, bool) {
	diameter := min(15, 8192/(depth+1)) + 1
	word := motion.WrappedPoint{X: int16(point.X), Y: int16(point.Y), Z: int16(depth)}
	p, projected := c.projection.Project(word, [3]int16{c.offset[0], c.offset[1], 0})
	x := int16((uint16(word.X)+uint16(c.offset[0]))&1023) - 512
	y := int16((uint16(word.Y)+uint16(c.offset[1]))&1023) - 512
	py := p.Y + (16-diameter)/2
	pose := contactPose{x: p.X, y: py - 26, frame: 16 - diameter, height: diameter,
		visible: projected && x != 0 && y != 0 && p.X >= 0 && p.X < 336 && py >= 0 && py < 302}
	c.poses[slot] = pose
	return sprites.FieldSample{X: float64(pose.x), Y: float64(pose.y), Z: float64(depth),
		Image: pose.frame, Scale: 1}, pose.visible
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
	velocity, ok := c.velocity.Sample(c.angles)
	if !ok {
		panic("demo: invalid contact velocity phase")
	}
	c.offset[0] = int16(uint16(c.offset[0]) + uint16(velocity[0]))
	c.offset[1] = int16(uint16(c.offset[1]) + uint16(velocity[1]))
	if err := c.queue.Step(-int(velocity[2])); err != nil {
		panic(err)
	}
	c.head, c.depth = c.queue.Head(), c.queue.Depth()
	return true
}

func contactVelocity(angles [3]int16, sines []int16) [3]int16 {
	velocity, err := motion.NewWordEulerVelocity(motion.WordEulerVelocityConfig{Sines: sines, Period: 720, Quantum: 2, Quarter: 180, OutputShift: [3]uint8{10, 10, 8}, PhasePolicy: motion.WordPhaseGuarded})
	if err != nil {
		panic(err)
	}
	v, ok := velocity.Sample(angles)
	if !ok {
		panic("demo: invalid contact velocity phase")
	}
	return v
}
