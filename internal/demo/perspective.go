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

type perspectiveEffect struct {
	clock        *perspectiveClock
	text         *scrolling.Scrolling
	white, layer *ebiten.Image
	fontPalette  *ebiten.Image
	batch        *render.Batch
	palette      *copperPalette
}

func newPerspective(data source.PerspectiveData, palette *copperPalette) (*perspectiveEffect, error) {
	e := &perspectiveEffect{clock: newPerspectiveClock(data), white: ebiten.NewImage(1, 1), layer: render.NewSurface(Width, Height), batch: render.NewBatch(1024), palette: palette}
	e.white.Fill(color.White)
	colors := image.NewNRGBA(image.Rect(0, 0, 1, Height))
	for y := 0; y < Height; y++ {
		level := 0
		for i, start := range []int{36, 39, 43, 47, 51, 59, 69, 80, 95} {
			if y >= start {
				level = i + 1
			}
		}
		colors.SetNRGBA(0, y, source.RGB12(uint16(level)))
	}
	e.fontPalette = ebiten.NewImageFromImage(colors)
	e.batch.Options.FillRule = ebiten.FillRuleEvenOdd
	glyphs := make([]scrolling.Glyph, 8)
	for i := range glyphs {
		glyphs[i] = scrolling.Glyph{Advance: 22}
	}
	var err error
	e.text, err = scrolling.New(scrolling.Config{Glyphs: glyphs})
	if err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}
func (e *perspectiveEffect) Update(kit.Frame) error { e.clock.Step(); return nil }
func (e *perspectiveEffect) paint(_ *ebiten.Image, s scrolling.Sample, _ ebiten.DrawImageOptions) {
	glyph := e.clock.data.Glyphs[e.clock.letters[s.Index]]
	vertex := func(p source.PolarGlyphPoint) ebiten.Vertex {
		v := e.clock.Point(s.Index, p)
		return render.Vertex(float64(v.X), float64(v.Y), 0.5, float64(v.Y), color.White)
	}
	for _, contour := range glyph.Contours {
		first := vertex(contour[0])
		for i := 1; i+1 < len(contour)-1; i++ {
			e.batch.Triangle(first, vertex(contour[i]), vertex(contour[i+1]))
		}
	}
}
func (e *perspectiveEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	if c.done {
		return
	}
	e.layer.Clear()
	view := e.layer.SubImage(image.Rect(0, 8, Width, 208)).(*ebiten.Image)
	if c.phase == 2 {
		e.batch.Begin(view, e.fontPalette)
		state := scrolling.IdentityState()
		state.Paint = e.paint
		e.text.DrawAt(view, state)
		e.batch.Flush()
	}
	// Source point colors use OR, independently of the outline parity mask.
	e.batch.Options.FillRule = ebiten.FillRuleFillAll
	e.batch.Begin(view, e.white)
	colors := [4]uint16{0, 0xfff, 0x68d, 0x359}
	for _, index := range c.touched {
		e.batch.Rect(float64(index%Width), float64(index/Width+8), 1, 1, image.Rect(0, 0, 1, 1), source.RGB12(colors[c.pointMasks[index]]))
	}
	e.batch.Flush()
	e.batch.Options.FillRule = ebiten.FillRuleEvenOdd
	e.palette.Draw(dst, e.layer, source.PaletteScaled32, c.level)
	border := source.RGB12(uint16(c.border))
	e.batch.Options.FillRule = ebiten.FillRuleFillAll
	e.batch.Begin(dst, e.white)
	e.batch.Rect(0, 7, Width, 1, image.Rect(0, 0, 1, 1), border)
	e.batch.Rect(0, 209, Width, 1, image.Rect(0, 0, 1, 1), border)
	e.batch.Flush()
	e.batch.Options.FillRule = ebiten.FillRuleEvenOdd
}
func (e *perspectiveEffect) Close() {
	if e.text != nil {
		e.text.Close()
	}
	e.white.Deallocate()
	e.layer.Deallocate()
	e.fontPalette.Deallocate()
}
