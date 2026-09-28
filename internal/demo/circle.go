package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// circleEffect binds the outline-font's polar painter to scrolling.New. Text
// transport and curve controls remain the authored byte-sized source program.
type circleEffect struct {
	clock                          *circleClock
	text                           *scrolling.Scrolling
	batch                          *render.Batch
	white, layer, mountain, raster *ebiten.Image
	palette                        *copperPalette
}

func newCircle(data source.CircleData, palette *copperPalette) (*circleEffect, error) {
	e := &circleEffect{clock: newCircleClock(data), batch: render.NewBatch(1024), white: ebiten.NewImage(1, 1),
		layer: render.NewSurface(Width, Height), mountain: ebiten.NewImageFromImage(data.Mountain), raster: ebiten.NewImageFromImage(data.Raster), palette: palette}
	e.white.Fill(color.White)
	e.batch.Options.FillRule = ebiten.FillRuleEvenOdd
	glyphs := make([]scrolling.Glyph, 8)
	for i := range glyphs {
		glyphs[i] = scrolling.Glyph{Advance: 30}
	}
	text, err := scrolling.New(scrolling.Config{Glyphs: glyphs})
	if err != nil {
		e.Close()
		return nil, err
	}
	e.text = text
	return e, nil
}

func (e *circleEffect) Update(kit.Frame) error { e.clock.Step(); return nil }

func (e *circleEffect) paint(_ *ebiten.Image, s scrolling.Sample, _ ebiten.DrawImageOptions) {
	glyph := e.clock.data.Glyphs[e.clock.letters[s.Index]]
	vertex := func(p source.PolarGlyphPoint) ebiten.Vertex {
		v := e.clock.Point(s.Index, p)
		return render.Vertex(float64(v.X), float64(v.Y), 0, 0, source.RGB12(0x3ad))
	}
	for _, contour := range glyph.Contours {
		first := vertex(contour[0])
		for i := 1; i+1 < len(contour)-1; i++ {
			e.batch.Triangle(first, vertex(contour[i]), vertex(contour[i+1]))
		}
	}
}

func (e *circleEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	if c.done {
		return
	}
	e.palette.Draw(dst, e.raster, source.PaletteScaled32, c.level)
	if c.active {
		e.layer.Clear()
		view := e.layer.SubImage(image.Rect(0, 8, Width, 208)).(*ebiten.Image)
		e.batch.Begin(view, e.white)
		state := scrolling.IdentityState()
		state.Paint = e.paint
		e.text.DrawAt(view, state)
		e.batch.Flush()
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, -2)
		op.ColorScale.Scale(0, 0.6, float32(11.0/13.0), 1)
		dst.DrawImage(e.layer, &op)
		dst.DrawImage(e.layer, nil)
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 8)
	dst.DrawImage(e.mountain, &op)
}

func (e *circleEffect) Close() {
	if e.text != nil {
		e.text.Close()
	}
	for _, img := range []*ebiten.Image{e.white, e.layer, e.mountain, e.raster} {
		if img != nil {
			img.Deallocate()
		}
	}
}
