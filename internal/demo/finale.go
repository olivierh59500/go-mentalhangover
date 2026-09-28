package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type finaleEffect struct {
	clock                  *finaleClock
	logo, logoLayer, floor *ebiten.Image
	balls                  []*ebiten.Image
	batch                  *render.Batch
	palette                *copperPalette
	pixels                 []byte
}

func newFinale(data source.FinaleData, palette *copperPalette) *finaleEffect {
	e := &finaleEffect{clock: newFinaleClock(data), logo: ebiten.NewImageFromImage(data.Logo), logoLayer: render.NewSurface(Width, Height),
		floor: render.NewSurface(Width, Height), batch: render.NewBatch(256), palette: palette, pixels: make([]byte, Width*Height*4)}
	for _, ball := range data.Balls {
		e.balls = append(e.balls, ebiten.NewImageFromImage(ball))
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(16, 14)
	e.logoLayer.DrawImage(e.logo, &op)
	return e
}

func (e *finaleEffect) Update(kit.Frame) error {
	e.clock.Step()
	c := e.clock
	clear(e.pixels)
	if c.tick >= 64 {
		for y := 166; y < Height; y++ {
			a, b := c.rowColors(y)
			first, second := source.RGB12(a), source.RGB12(b)
			row := y - 138
			for x := 0; x < Width; x++ {
				word := c.data.Columns[row*48+x/8]
				color := first
				if word>>uint(7-x%8)&1 != 0 {
					color = second
				}
				i := (y*Width + x) * 4
				e.pixels[i], e.pixels[i+1], e.pixels[i+2], e.pixels[i+3] = color.R, color.G, color.B, 255
			}
		}
	}
	e.floor.WritePixels(e.pixels)
	return nil
}

func (e *finaleEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	e.palette.Draw(dst, e.logoLayer, source.PaletteScaled64, c.logoLevel)
	dst.DrawImage(e.floor, nil)
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
func (e *finaleEffect) Close() {
	for _, img := range append(e.balls, e.logo, e.logoLayer, e.floor) {
		img.Deallocate()
	}
}
