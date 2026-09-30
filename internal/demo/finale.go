package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type finaleEffect struct {
	clock                    *finaleClock
	logo, logoLayer, columns *ebiten.Image
	floor                    *composite.PaletteGrid
	balls                    []*ebiten.Image
	slots                    *sprites.ImageSlots
	palette                  *copperPalette
	colors                   [2][]color.NRGBA
}

func newFinale(data source.FinaleData, palette *copperPalette) (*finaleEffect, error) {
	floor, err := composite.NewPaletteGrid(composite.PaletteGridConfig{
		Width: Width, Height: Height, Columns: 1, Rows: Height, ControlChannel: composite.BitplaneRed,
	})
	if err != nil {
		return nil, err
	}
	mask := image.NewNRGBA(image.Rect(0, 0, Width, Height))
	for y := 166; y < Height; y++ {
		for x := 0; x < Width; x++ {
			if data.Columns[(y-138)*48+x/8]>>uint(7-x%8)&1 != 0 {
				mask.SetNRGBA(x, y, color.NRGBA{255, 255, 255, 255})
			}
		}
	}
	e := &finaleEffect{clock: newFinaleClock(data), logo: ebiten.NewImageFromImage(data.Logo), logoLayer: render.NewSurface(Width, Height),
		columns: ebiten.NewImageFromImage(mask), floor: floor, palette: palette,
		colors: [2][]color.NRGBA{make([]color.NRGBA, Height), make([]color.NRGBA, Height)}}
	for _, ball := range data.Balls {
		e.balls = append(e.balls, ebiten.NewImageFromImage(ball))
	}
	e.slots, err = sprites.NewImageSlots(sprites.ImageSlotsConfig{Images: e.balls, MaxSlots: 24, Select: e.selectBalls})
	if err != nil {
		e.Close()
		return nil, err
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(16, 14)
	e.logoLayer.DrawImage(e.logo, &op)
	return e, nil
}

func (e *finaleEffect) Update(f kit.Frame) error {
	e.clock.Step()
	c := e.clock
	clear(e.colors[0])
	clear(e.colors[1])
	if c.tick >= 64 {
		for y := 166; y < Height; y++ {
			a, b := c.rowColors(y)
			first, second := source.RGB12(a), source.RGB12(b)
			e.colors[0][y] = color.NRGBA{R: first.R, G: first.G, B: first.B, A: 255}
			e.colors[1][y] = color.NRGBA{R: second.R, G: second.G, B: second.B, A: 255}
		}
	}
	// Only a two-color row bank changes. The original column mask stays on the
	// GPU, avoiding a full-screen CPU raster and upload on every PAL update.
	if err := e.floor.SetColors(e.colors[0], e.colors[1]); err != nil {
		return err
	}
	return e.slots.Update(f)
}

func (e *finaleEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	e.palette.Draw(dst, e.logoLayer, source.PaletteScaled64, c.logoLevel)
	e.drawFloor(dst)
	e.slots.Draw(dst)
}

// selectBalls supplies the authored crop and lower border; DCK owns the slot
// window, sprite ordering, texture switches and triangle submission.
func (e *finaleEffect) selectBalls(_ kit.Frame, slots []sprites.ImageSlot) (int, error) {
	c := e.clock
	for i := 0; i < c.ballCount; i++ {
		p := c.poses[i]
		h := min(p.size, 278-p.y)
		slots[i] = sprites.ImageSlot{Hidden: !p.visible || p.material >= len(e.balls) || h <= 0}
		if slots[i].Hidden {
			continue
		}
		x := (32 - p.size) * 32
		slots[i].Image = p.material
		slots[i].X, slots[i].Y = float64(p.x), float64(p.y)
		slots[i].Width, slots[i].Height = 32, float64(h)
		slots[i].Source = image.Rect(x, 0, x+32, h)
	}
	return c.ballCount, nil
}

func (e *finaleEffect) drawFloor(dst *ebiten.Image) {
	if err := e.floor.Draw(dst, e.columns, nil, composite.PaletteGridState{}); err != nil {
		panic(err)
	}
}

func (e *finaleEffect) Close() {
	if e.slots != nil {
		e.slots.Close()
	}
	for _, img := range append(e.balls, e.logo, e.logoLayer, e.columns) {
		img.Deallocate()
	}
	e.floor.Close()
}
