package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type finaleEffect struct {
	clock                    *finaleClock
	logo, logoLayer, columns *ebiten.Image
	floor                    *composite.PaletteGrid
	balls                    []*ebiten.Image
	batch                    *render.Batch
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
		columns: ebiten.NewImageFromImage(mask), floor: floor, batch: render.NewBatch(256), palette: palette,
		colors: [2][]color.NRGBA{make([]color.NRGBA, Height), make([]color.NRGBA, Height)}}
	for _, ball := range data.Balls {
		e.balls = append(e.balls, ebiten.NewImageFromImage(ball))
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(16, 14)
	e.logoLayer.DrawImage(e.logo, &op)
	return e, nil
}

func (e *finaleEffect) Update(kit.Frame) error {
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
	return e.floor.SetColors(e.colors[0], e.colors[1])
}

func (e *finaleEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	e.palette.Draw(dst, e.logoLayer, source.PaletteScaled64, c.logoLevel)
	e.drawFloor(dst)
	// The source shifts its 24-entry depth queue; near balls draw last.
	for i := 0; i < c.ballCount; i++ {
		p := c.poses[i]
		if !p.visible || p.material >= len(e.balls) {
			continue
		}
		xOffset := (32 - p.size) * 32
		e.batch.Begin(dst, e.balls[p.material])
		height := min(p.size, 278-p.y)
		if height <= 0 {
			continue
		}
		e.batch.Rect(float64(p.x), float64(p.y), 32, float64(height), image.Rect(xOffset, 0, xOffset+32, height), color.White)
		e.batch.Flush()
	}
}

func (e *finaleEffect) drawFloor(dst *ebiten.Image) {
	if err := e.floor.Draw(dst, e.columns, nil, composite.PaletteGridState{}); err != nil {
		panic(err)
	}
}

func (e *finaleEffect) Close() {
	for _, img := range append(e.balls, e.logo, e.logoLayer, e.columns) {
		img.Deallocate()
	}
	e.floor.Close()
}
