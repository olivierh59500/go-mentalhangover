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
	clock                          *finaleClock
	logo, logoLayer, columns, rows *ebiten.Image
	floorShader                    *ebiten.Shader
	balls                          []*ebiten.Image
	batch                          *render.Batch
	palette                        *copperPalette
	pixels                         []byte
	floorVertices                  [4]ebiten.Vertex
}

const finaleFloorShader = `//kage:unit pixels
package main
func Fragment(dst vec4, src vec2, color vec4) vec4 {
    y := floor((src - imageSrc0Origin()).y) + 0.5
    a := imageSrc1At(imageSrc0Origin() + vec2(0.5, y))
    b := imageSrc1At(imageSrc0Origin() + vec2(1.5, y))
    return mix(a, b, imageSrc0At(src).r)
}
`

func newFinale(data source.FinaleData, palette *copperPalette) (*finaleEffect, error) {
	shader, err := ebiten.NewShader([]byte(finaleFloorShader))
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
		columns: ebiten.NewImageFromImage(mask), rows: render.NewSurface(2, Height), floorShader: shader,
		batch: render.NewBatch(256), palette: palette, pixels: make([]byte, 2*Height*4)}
	e.floorVertices = [4]ebiten.Vertex{render.Vertex(0, 0, 0, 0, color.White), render.Vertex(Width, 0, Width, 0, color.White),
		render.Vertex(0, Height, 0, Height, color.White), render.Vertex(Width, Height, Width, Height, color.White)}
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
	clear(e.pixels)
	if c.tick >= 64 {
		for y := 166; y < Height; y++ {
			a, b := c.rowColors(y)
			first, second := source.RGB12(a), source.RGB12(b)
			i := y * 8
			e.pixels[i], e.pixels[i+1], e.pixels[i+2], e.pixels[i+3] = first.R, first.G, first.B, 255
			e.pixels[i+4], e.pixels[i+5], e.pixels[i+6], e.pixels[i+7] = second.R, second.G, second.B, 255
		}
	}
	// Only a two-color row bank changes. The original column mask stays on the
	// GPU, avoiding a full-screen CPU raster and upload on every PAL update.
	e.rows.WritePixels(e.pixels)
	return nil
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
	op := ebiten.DrawTrianglesShaderOptions{Images: [4]*ebiten.Image{e.columns, e.rows}}
	dst.DrawTrianglesShader(e.floorVertices[:], []uint16{0, 1, 2, 1, 3, 2}, e.floorShader, &op)
}

func (e *finaleEffect) Close() {
	for _, img := range append(e.balls, e.logo, e.logoLayer, e.columns, e.rows) {
		img.Deallocate()
	}
	e.floorShader.Deallocate()
}
