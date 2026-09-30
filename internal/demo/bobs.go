package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// bobEffect submits the original projected copies through one retained DCK
// batch. A three-column source texture preserves copper colors by output row.
type bobEffect struct {
	clock              *bobClock
	models             []source.SolidModel
	points             []source.Point2
	fontImage, palette *ebiten.Image
	text               *scrolling.Scrolling
	batch              *render.Batch
	ready              bool
	faces              []clippedFace
}

func newBOBEffect(data []byte) (*bobEffect, error) {
	decoded, err := source.ReadBOBs(data)
	if err != nil {
		return nil, err
	}
	models, err := source.BOBModels(data)
	if err != nil {
		return nil, err
	}
	effect := &bobEffect{clock: newBOBClock(decoded), models: models, points: make([]source.Point2, 8), batch: render.NewBatch(1024)}
	effect.faces = make([]clippedFace, 0, 6)
	effect.fontImage = ebiten.NewImageFromImage(decoded.Font)
	effect.palette = ebiten.NewImageFromImage(decoded.Palette)
	glyphs := make(map[rune]font.Glyph, 59)
	for i, advance := range decoded.Advances {
		glyphs[rune(i+32)] = font.Glyph{Rect: image.Rect(i*16, 0, i*16+16, 14), Advance: float64(advance)}
	}
	glyphs[' '] = font.Glyph{Advance: float64(decoded.Advances[0])}
	metrics, err := font.New(font.Config{Bounds: decoded.Font.Bounds(), Glyphs: glyphs, LineHeight: 14, SpaceAdvance: float64(decoded.Advances[0])})
	if err != nil {
		effect.Close()
		return nil, err
	}
	effect.text, err = scrolling.New(scrolling.Config{Text: decoded.Text, Fonts: map[string]scrolling.Face{"default": {Atlas: effect.fontImage, Metrics: metrics}}})
	if err != nil {
		effect.Close()
		return nil, err
	}
	return effect, nil
}

func (e *bobEffect) Update(kit.Frame) error {
	e.ready = e.clock.Step()
	if !e.ready || e.clock.depth == 4100 {
		return nil
	}
	model := e.models[e.clock.model]
	if err := projectBOB(model.Points, solidMatrix(e.clock.angles, e.clock.data.Sines), e.clock.depth, e.points[:len(model.Points)]); err != nil {
		return err
	}
	e.faces = e.faces[:0]
	for _, face := range model.Faces {
		a, b, c := e.points[face.Vertices[0]], e.points[face.Vertices[1]], e.points[face.Vertices[2]]
		if int32(b.X-a.X)*int32(c.Y-a.Y)-int32(b.Y-a.Y)*int32(c.X-a.X) < 0 {
			continue
		}
		var points [8]polygonPoint
		for i, index := range face.Vertices {
			p := e.points[index]
			points[i] = polygonPoint{float64(p.X), float64(p.Y)}
		}
		clipped := clipBOB(points[:len(face.Vertices)])
		clipped.material = face.Material
		if clipped.count >= 3 {
			e.faces = append(e.faces, clipped)
		}
	}
	return nil
}

func (e *bobEffect) Draw(dst *ebiten.Image) {
	if !e.ready {
		return
	}
	c := e.clock
	state := scrolling.IdentityState()
	state.First, state.End = c.first, c.fetched
	state.Y = 15
	state.Map = func(sample scrolling.Sample, options *ebiten.DrawImageOptions) bool {
		options.GeoM.Translate(float64(c.origins[sample.Index]-c.distance)-sample.X, 0)
		return true
	}
	e.text.DrawAt(dst, state)
	if c.depth == 4100 || c.count == 0 {
		return
	}
	e.batch.Begin(dst, e.palette)
	for i := 0; i < c.count; i++ {
		x, y := c.Offset(i)
		for _, face := range e.faces {
			vertex := func(point polygonPoint) ebiten.Vertex {
				px, py := float64(x)+point.x, float64(y)+point.y
				return render.Vertex(px, py, float64(face.material)-0.5, py, color.White)
			}
			e.batch.Fan(face.count, func(j int) ebiten.Vertex { return vertex(face.points[j]) })
		}
	}
	e.batch.Flush()
}

func (e *bobEffect) Close() {
	if e.text != nil {
		e.text.Close()
	}
	if e.fontImage != nil {
		e.fontImage.Deallocate()
	}
	if e.palette != nil {
		e.palette.Deallocate()
	}
}
