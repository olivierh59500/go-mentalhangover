package demo

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// filledSolidEffect shares source geometry/projection and DCK batching with the
// BOB unit. Its copper palette atlas retains all 54 phases without GPU uploads.
type filledSolidEffect struct {
	data    source.FilledSolid
	clock   *vectorClock
	points  []source.Point2
	batch   *render.Batch
	palette *ebiten.Image
	phase   int
	ready   bool
}

func newFilledSolid(data source.FilledSolid) *filledSolidEffect {
	return &filledSolidEffect{data: data, clock: newVectorClock(source.VectorModel{Initial: data.Initial, Cues: data.Cues}),
		points: make([]source.Point2, len(data.Model.Points)), batch: render.NewBatch(256), palette: ebiten.NewImageFromImage(data.Palette)}
}

func (e *filledSolidEffect) Update(kit.Frame) error {
	e.ready = e.clock.Step()
	if !e.ready {
		return nil
	}
	e.phase = (e.phase + 1) % 54
	s := e.clock.state
	return projectSolid(e.data.Model.Points, solidMatrix([3]int16{s[0], s[1], s[2]}, e.data.Sines), s, e.points)
}

func (e *filledSolidEffect) Draw(dst *ebiten.Image) {
	if !e.ready {
		return
	}
	e.batch.Begin(dst, e.palette)
	for _, face := range e.data.Model.Faces {
		a, b, c := e.points[face.Vertices[0]], e.points[face.Vertices[1]], e.points[face.Vertices[2]]
		if int32(b.X-a.X)*int32(c.Y-a.Y)-int32(b.Y-a.Y)*int32(c.X-a.X) < 0 {
			continue
		}
		vertex := func(p source.Point2) ebiten.Vertex {
			x, y := float64(p.X)-176, float64(p.Y)-1
			return render.Vertex(x, y, float64(e.phase*3)+float64(face.Material)-0.5, y, color.White)
		}
		e.batch.Fan(len(face.Vertices), func(i int) ebiten.Vertex { return vertex(e.points[face.Vertices[i]]) })
	}
	e.batch.Flush()
}

func (e *filledSolidEffect) Close() {
	if e.palette != nil {
		e.palette.Deallocate()
	}
}
