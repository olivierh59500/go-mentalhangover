package demo

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// vectorEffect projects the source's word-sized points, then delegates bounded
// triangle submission to DCK. One even-odd batch preserves concave contours,
// holes and the sign's crossing edges without replacing them with solid fans.
type vectorEffect struct {
	model  source.VectorModel
	clock  *vectorClock
	sines  []int16
	points []source.Point2
	batch  *render.Batch
	white  *ebiten.Image
	color  color.NRGBA
	ready  bool
}

func newVectorEffect(model source.VectorModel, sines []int16) (*vectorEffect, error) {
	if len(sines) != 2048 || len(model.Cues) == 0 {
		return nil, fmt.Errorf("vector effect requires original clocks and waves")
	}
	effect := &vectorEffect{model: model, clock: newVectorClock(model), sines: sines,
		points: make([]source.Point2, len(model.Points)), batch: render.NewBatch(512),
		white: ebiten.NewImage(1, 1)}
	effect.white.Fill(color.White)
	effect.batch.Options.FillRule = ebiten.FillRuleEvenOdd
	return effect, nil
}

func (effect *vectorEffect) Update(kit.Frame) error {
	if !effect.clock.Step() {
		effect.ready = false
		return nil
	}
	state := effect.clock.state
	if err := projectVector(effect.model.Points, vectorMatrix(state, effect.sines), state, effect.points); err != nil {
		return err
	}
	// The palette table at 0x1b248 runs from full 0x24f to zero as depth grows.
	level := 16 - (max(2048, min(4096, int(state[5])))-2048)/128
	word := uint16((2*level/16)<<8 | (4*level/16)<<4 | 15*level/16)
	effect.color = source.RGB12(word)
	effect.ready = true
	return nil
}

func (effect *vectorEffect) Draw(dst *ebiten.Image) {
	if !effect.ready {
		return
	}
	effect.batch.Begin(dst, effect.white)
	for _, contour := range effect.model.Contours {
		indices := contour.Indices[:len(contour.Indices)-1]
		origin := effect.points[indices[0]]
		first := render.Vertex(float64(origin.X), float64(origin.Y), 0, 0, effect.color)
		for i := 1; i+1 < len(indices); i++ {
			b, c := effect.points[indices[i]], effect.points[indices[i+1]]
			effect.batch.Triangle(first,
				render.Vertex(float64(b.X), float64(b.Y), 0, 0, effect.color),
				render.Vertex(float64(c.X), float64(c.Y), 0, 0, effect.color))
		}
	}
	effect.batch.Flush()
}

func (effect *vectorEffect) Close() {
	if effect.white != nil {
		effect.white.Deallocate()
		effect.white = nil
	}
}
