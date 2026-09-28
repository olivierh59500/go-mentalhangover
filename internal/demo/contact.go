package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type contactEffect struct {
	clock               *contactClock
	balls, title, layer *ebiten.Image
	batch               *render.Batch
	palette             *copperPalette
}

func newContact(data source.ContactData, palette *copperPalette) *contactEffect {
	return &contactEffect{clock: newContactClock(data), balls: ebiten.NewImageFromImage(data.Balls),
		title: ebiten.NewImageFromImage(data.Title), layer: render.NewSurface(Width, Height), batch: render.NewBatch(1024), palette: palette}
}

func (e *contactEffect) Update(kit.Frame) error { e.clock.Step(); return nil }
func (e *contactEffect) Draw(dst *ebiten.Image) {
	if e.clock.done {
		return
	}
	e.layer.Clear()
	view := e.layer.SubImage(image.Rect(16, 0, 336, Height)).(*ebiten.Image)
	e.batch.Begin(view, e.balls)
	for _, p := range e.clock.poses {
		if !p.visible {
			continue
		}
		e.batch.Rect(float64(p.x), float64(p.y), 16, float64(p.height), image.Rect(p.frame*16, 0, p.frame*16+16, p.height), color.White)
	}
	e.batch.Flush()
	e.palette.Draw(dst, e.layer, source.PaletteScaled32, e.clock.ballLevel)
	if e.clock.titleLevel > 0 {
		e.palette.Draw(dst, e.title, source.PaletteScaled32, e.clock.titleLevel)
	}
}
func (e *contactEffect) Close() {
	for _, img := range []*ebiten.Image{e.balls, e.title, e.layer} {
		img.Deallocate()
	}
}
