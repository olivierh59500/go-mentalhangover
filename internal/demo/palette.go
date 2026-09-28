package demo

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// The source artwork contains opaque 12-bit palette colors and transparent
// index zero. Copper transitions modify each nibble with integer truncation.
const paletteShaderSource = `//kage:unit pixels
package main
var Mode float
var Level float
func Fragment(dst vec4, src vec2, color vec4) vec4 {
    c := imageSrc0At(src)
    if c.a < 0.5 { return vec4(0) }
    target := floor(c.rgb * 15 + vec3(0.5))
    value := target
    if Mode == 1 { value = vec3(Level) }
    if Mode == 2 { value = vec3(15) - floor((vec3(15) - target) * Level / 32) }
    if Mode == 3 { value = floor(target * Level / 16) }
    return vec4(value / 15, c.a)
}
`

type copperPalette struct {
	shader   *ebiten.Shader
	uniforms map[string]any
}

func newCopperPalette() (*copperPalette, error) {
	shader, err := ebiten.NewShader([]byte(paletteShaderSource))
	if err != nil {
		return nil, err
	}
	return &copperPalette{shader: shader, uniforms: map[string]any{"Mode": float32(0), "Level": float32(0)}}, nil
}

func (palette *copperPalette) Draw(dst, src *ebiten.Image, mode source.PaletteMode, level int) {
	if mode == source.PaletteTarget {
		dst.DrawImage(src, nil)
		return
	}
	palette.uniforms["Mode"] = float32(mode)
	palette.uniforms["Level"] = float32(level)
	options := ebiten.DrawRectShaderOptions{Images: [4]*ebiten.Image{src}, Uniforms: palette.uniforms}
	dst.DrawRectShader(src.Bounds().Dx(), src.Bounds().Dy(), palette.shader, &options)
}

func (palette *copperPalette) Close() { palette.shader.Deallocate() }
