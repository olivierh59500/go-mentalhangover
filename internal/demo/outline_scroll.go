package demo

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

const outlineMode = "contours"

// newOutlineScroll adapts decoded byte-pair artwork and authored placement to
// the shared contour-font mode. DCK owns the visible slots and geometry buffers.
func newOutlineScroll(art map[byte]source.PolarGlyph, advance float64, letter func(int) byte,
	point func(int, source.PolarGlyphPoint) source.Point2, texture *ebiten.Image, paint *color.NRGBA) (*scrolling.Scrolling, error) {
	glyphs := make(map[rune]font.ContourGlyph, len(art))
	for character, glyph := range art {
		converted := font.ContourGlyph{Advance: advance, Contours: make([][]geometry.Vec2, len(glyph.Contours))}
		for i, contour := range glyph.Contours {
			for _, p := range contour {
				converted.Contours[i] = append(converted.Contours[i], geometry.Vec2{X: float64(p.Angle), Y: float64(p.Row)})
			}
		}
		glyphs[rune(character)] = converted
	}
	bank, err := font.NewContourBank(font.ContourBankConfig{Glyphs: glyphs, Closed: true})
	if err != nil {
		return nil, err
	}
	config := scrolling.ContourPainterConfig{Fonts: map[string]*font.ContourBank{"main": bank}, Font: "main",
		Texture: texture, Color: paint, FillRule: ebiten.FillRuleEvenOdd, BatchTriangles: 1024,
		Map: func(s scrolling.ContourSample, _ ebiten.GeoM) (geometry.Vec2, bool) {
			p := point(s.Index, source.PolarGlyphPoint{Angle: byte(s.Point.X), Row: byte(s.Point.Y)})
			return geometry.Vec2{X: float64(p.X), Y: float64(p.Y)}, true
		}}
	if texture != nil {
		config.UV = func(_ scrolling.ContourSample, p geometry.Vec2) geometry.Vec2 { return geometry.Vec2{X: .5, Y: p.Y} }
	}
	return scrolling.New(scrolling.Config{Shape: outlineMode,
		GlyphWindow: &scrolling.GlyphWindowConfig{Count: 8, Advance: advance,
			Glyph: func(slot int) scrolling.Glyph { return scrolling.Glyph{Rune: rune(letter(slot)), Font: "main"} }},
		Modes: map[string]scrolling.Mode{outlineMode: {Contours: &config}},
	})
}
