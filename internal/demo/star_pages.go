package demo

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type starPageEffect struct {
	clock     *starPageClock
	pages     []*ebiten.Image
	fontImage *ebiten.Image
	palette   *copperPalette
}

func newStarPages(data source.StarPages, palette *copperPalette) (*starPageEffect, error) {
	e := &starPageEffect{clock: newStarPageClock(data), palette: palette, fontImage: ebiten.NewImageFromImage(data.Font)}
	glyphs := make(map[rune]font.Glyph, 59)
	for i, a := range data.Advances {
		glyphs[rune(i+32)] = font.Glyph{Rect: image.Rect(i*16, 0, i*16+16, 14), Advance: float64(a)}
	}
	glyphs[' '] = font.Glyph{Advance: float64(data.Advances[0])}
	metrics, err := font.New(font.Config{Bounds: data.Font.Bounds(), Glyphs: glyphs, LineHeight: 16, SpaceAdvance: float64(data.Advances[0])})
	if err != nil {
		e.Close()
		return nil, err
	}
	for _, card := range data.Cards {
		text, err := scrolling.New(scrolling.Config{Text: strings.Join(card.Lines, "\n"), Y: float64(card.Y) - 10,
			Fonts: map[string]scrolling.Face{"default": {Atlas: e.fontImage, Metrics: metrics}},
			Page:  &scrolling.PageConfig{Width: Width, LineHeight: 16, Align: scrolling.AlignCenter},
			Map: func(s scrolling.Sample, op *ebiten.DrawImageOptions) bool {
				op.GeoM.Translate(math.Ceil(s.X)-s.X, 0)
				return true
			}})
		if err != nil {
			e.Close()
			return nil, err
		}
		page := render.NewSurface(Width, Height)
		text.Draw(page)
		text.Close()
		e.pages = append(e.pages, page)
	}
	return e, nil
}

func (e *starPageEffect) Update(kit.Frame) error { e.clock.Step(); return nil }

func (e *starPageEffect) Draw(dst *ebiten.Image) {
	c := e.clock
	if c.frame < 0 || c.done {
		return
	}
	var colors [4]color.NRGBA
	for i, word := range c.data.StarColors {
		colors[i+1] = source.RGB12(source.PaletteWord(word, source.PaletteScaled32, c.starLevel))
	}
	if err := c.plane.SetPalette(colors[:]); err != nil {
		panic(err)
	}
	c.plane.Draw(dst)
	if c.page >= 0 {
		e.palette.Draw(dst, e.pages[c.page], source.PaletteScaled32, c.textLevel)
	}
}

func (e *starPageEffect) Close() {
	for _, page := range e.pages {
		page.Deallocate()
	}
	e.clock.plane.Close()
	if e.fontImage != nil {
		e.fontImage.Deallocate()
	}
}
