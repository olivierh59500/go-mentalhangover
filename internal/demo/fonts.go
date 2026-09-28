package demo

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func serifAtlas(resident []byte) (*scrolling.Atlas, *ebiten.Image, error) {
	pixels, advances, err := source.SerifFont(resident)
	if err != nil {
		return nil, nil, err
	}
	glyphs := make(map[rune]font.Glyph, len(advances))
	for i, advance := range advances {
		glyphs[rune(i+32)] = font.Glyph{Rect: image.Rect(i*48, 0, i*48+48, 23), Advance: float64(advance)}
	}
	glyphs[' '] = font.Glyph{Advance: float64(advances[0])}
	metrics, err := font.New(font.Config{Bounds: pixels.Bounds(), Glyphs: glyphs,
		LineHeight: 24, SpaceAdvance: float64(advances[0]), Uppercase: true})
	if err != nil {
		return nil, nil, err
	}
	image := ebiten.NewImageFromImage(pixels)
	atlas, err := scrolling.NewAtlas(image, metrics)
	if err != nil {
		image.Deallocate()
		return nil, nil, err
	}
	return atlas, image, nil
}
