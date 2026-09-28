package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type patternedEffect struct {
	data     source.PatternedSolid
	clock    *vectorClock
	textures []*ebiten.Image
	points   []source.Point2
	batch    *render.Batch
	frame    int
	ready    bool
}

func newPatterned(data source.PatternedSolid) *patternedEffect {
	e := &patternedEffect{data: data, clock: newVectorClock(source.VectorModel{Initial: data.Initial, Cues: data.Cues}),
		points: make([]source.Point2, len(data.Points)), batch: render.NewBatch(256)}
	e.batch.Options.FillRule = ebiten.FillRuleEvenOdd
	e.batch.Options.Address = ebiten.AddressRepeat
	for _, texture := range data.Textures {
		e.textures = append(e.textures, ebiten.NewImageFromImage(texture))
	}
	// Source pixels are now retained on the GPU; only choreography stays in Go.
	e.data.Textures = nil
	return e
}

func (e *patternedEffect) Update(kit.Frame) error {
	advance := e.frame%e.data.Period == 0
	e.frame++
	if !advance {
		return nil
	}
	e.ready = e.clock.Step()
	if !e.ready {
		return nil
	}
	for i := 0; i < 3; i++ {
		angle := e.clock.state[i]
		if angle > 718 {
			angle -= 720
		} else if angle < 0 {
			angle += 720
		}
		e.clock.state[i] = angle
	}
	s := e.clock.state
	return projectPatterned(e.data.Points, patternedMatrix([3]int16{s[0], s[1], s[2]}, e.data.Sines), s, e.points)
}

func (e *patternedEffect) Level() int {
	depth := max(548, min(2590, int(e.clock.state[5])))
	return 127 - ((depth-548)&0xff0)/16
}

func (e *patternedEffect) Draw(dst *ebiten.Image) {
	if !e.ready {
		return
	}
	view := dst.SubImage(image.Rect(0, 30, Width, 227)).(*ebiten.Image)
	for _, face := range e.data.Faces {
		indices := face.Contours[0]
		a, b, c := e.points[indices[0]], e.points[indices[1]], e.points[indices[2]]
		if int32(b.X-a.X)*int32(c.Y-a.Y)-int32(b.Y-a.Y)*int32(c.X-a.X) < 0 {
			continue
		}
		minX, minY, maxX, maxY := 32767, 32767, -32768, -32768
		for _, contour := range face.Contours {
			for _, index := range contour {
				p := e.points[index]
				minX = min(minX, int(p.X))
				maxX = max(maxX, int(p.X))
				minY = min(minY, int(p.Y))
				maxY = max(maxY, int(p.Y))
			}
		}
		centerX, centerY := minX+(maxX-minX)/2, minY+(maxY-minY)/2
		vertex := func(p source.Point2) ebiten.Vertex {
			return render.Vertex(float64(p.X)-80, float64(p.Y)+27,
				float64(int(p.X)-centerX+128), float64(int(p.Y)-centerY+128), color.White)
		}
		e.batch.Begin(view, e.textures[int(face.Material)-1])
		for _, contour := range face.Contours {
			first := vertex(e.points[contour[0]])
			for i := 1; i+1 < len(contour)-1; i++ {
				e.batch.Triangle(first, vertex(e.points[contour[i]]), vertex(e.points[contour[i+1]]))
			}
		}
		e.batch.Flush()
	}
}

func (e *patternedEffect) Close() {
	for _, texture := range e.textures {
		texture.Deallocate()
	}
}
