package demo

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// Source operation order and integer divisors are authored parameters.
var copperStates = [...]composite.QuantizedColorState{
	{Mode: composite.QuantizedPassthrough},
	{Mode: composite.QuantizedReplace},
	{Mode: composite.QuantizedFromTarget, Target: [3]uint16{15, 15, 15}, Denominator: 32},
	{Mode: composite.QuantizedScale, Denominator: 16},
	{Mode: composite.QuantizedScale, Denominator: 128},
	{Mode: composite.QuantizedScale, Denominator: 32},
	{Mode: composite.QuantizedScale, Denominator: 64},
}

type copperPalette struct{ renderer *composite.QuantizedColor }

func newCopperPalette() (*copperPalette, error) {
	renderer, err := composite.NewQuantizedColor(composite.QuantizedColorConfig{AlphaThreshold: 0.5})
	if err != nil {
		return nil, err
	}
	return &copperPalette{renderer: renderer}, nil
}

func (palette *copperPalette) Draw(dst, src *ebiten.Image, mode source.PaletteMode, level int) {
	state := composite.QuantizedColorState{Mode: composite.QuantizedKeep}
	if int(mode) < len(copperStates) {
		state = copperStates[mode]
	}
	if state.Mode == composite.QuantizedReplace {
		value := uint16(max(0, min(15, level)))
		state.Target = [3]uint16{value, value, value}
	} else if state.Denominator != 0 {
		limit := int(state.Denominator)
		if mode == source.PaletteScaled128 {
			limit--
		}
		state.Numerator = uint32(max(0, min(limit, level)))
	}
	if err := palette.renderer.Draw(dst, src, state); err != nil {
		panic(err)
	}
}

func (palette *copperPalette) Close() { palette.renderer.Close() }
