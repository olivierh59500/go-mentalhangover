package demo

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type contactEffect struct {
	clock               *contactClock
	balls, title, layer *ebiten.Image
	palette             *copperPalette
}

func newContact(data source.ContactData, palette *copperPalette) *contactEffect {
	e := &contactEffect{clock: newContactClock(data), balls: ebiten.NewImageFromImage(data.Balls),
		title: ebiten.NewImageFromImage(data.Title), layer: render.NewSurface(Width, Height), palette: palette}
	frames := make([]image.Rectangle, 16)
	for frame := range frames {
		frames[frame] = image.Rect(frame*16, 0, frame*16+16, 16-frame)
	}
	e.clock.queue.Style = sprites.FieldStyle{Image: e.balls, Frames: frames}
	return e
}

func (e *contactEffect) Update(kit.Frame) error { e.clock.Step(); return nil }
func (e *contactEffect) Draw(dst *ebiten.Image) {
	if e.clock.done {
		return
	}
	e.layer.Clear()
	view := e.layer.SubImage(image.Rect(16, 0, 336, Height)).(*ebiten.Image)
	e.clock.queue.Draw(view)
	e.palette.Draw(dst, e.layer, source.PaletteScaled32, e.clock.ballLevel)
	if e.clock.titleLevel > 0 {
		e.palette.Draw(dst, e.title, source.PaletteScaled32, e.clock.titleLevel)
	}
}
func (e *contactEffect) Close() {
	e.clock.queue.Close()
	for _, img := range []*ebiten.Image{e.balls, e.title, e.layer} {
		img.Deallocate()
	}
}
