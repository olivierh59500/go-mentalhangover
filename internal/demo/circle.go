package demo

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// circleEffect supplies decoded artwork, curve controls and layer composition
// to DCK's shared byte-window transport and contour-font scrolling mode.
type circleEffect struct {
	clock                   *circleClock
	text                    *scrolling.Scrolling
	layer, mountain, raster *ebiten.Image
	palette                 *copperPalette
}

func newCircle(data source.CircleData, palette *copperPalette) (*circleEffect, error) {
	e := &circleEffect{clock: newCircleClock(data),
		layer: render.NewSurface(Width, Height), mountain: ebiten.NewImageFromImage(data.Mountain), raster: ebiten.NewImageFromImage(data.Raster), palette: palette}
	paint := source.RGB12(0x3ad)
	text, err := newOutlineScroll(data.Glyphs, 30, func(slot int) byte { return e.clock.letters[slot] }, e.clock.Point, nil, &paint)
	if err != nil {
		e.Close()
		return nil, err
	}
	e.text = text
	return e, nil
}

func (e *circleEffect) Update(f kit.Frame) error {
	if err := e.text.Err(); err != nil {
		return err
	}
	e.clock.Step()
	return e.text.Update(f)
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
		e.text.Draw(view)
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
	for _, img := range []*ebiten.Image{e.layer, e.mountain, e.raster} {
		if img != nil {
			img.Deallocate()
		}
	}
}
