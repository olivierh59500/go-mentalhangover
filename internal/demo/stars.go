package demo

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// originalStars binds the original 135 sprite headers to DCK's fixed-step field.
// Their horizontal bytes wrap independently at 256 hardware clocks; no random
// population, perspective projection or frame-rate-dependent motion is added.
func originalStars(resident []byte) (*sprites.AnimatedField, []*ebiten.Image, error) {
	const start = 0x9c6 - 0x21e
	if len(resident) < start+3*0x16c {
		return nil, nil, fmt.Errorf("missing original sprite bank")
	}
	positions := make([]motion.FrameParticle, 135)
	var images []*ebiten.Image
	palette := []uint16{0, 0x227, 0x449, 0x338}
	for bank := 0; bank < 3; bank++ {
		for i := 0; i < 45; i++ {
			header := resident[start+bank*0x16c+i*8:]
			position, control := binary.BigEndian.Uint16(header), binary.BigEndian.Uint16(header[2:])
			pixels, err := source.DecodePlanar(header[4:8], source.Planar{Width: 16, Height: 1, RowStride: 2,
				PlaneOffsets: []int{0, 2}, Palette: palette, TransparentZero: true})
			if err != nil {
				return nil, nil, err
			}
			images = append(images, ebiten.NewImageFromImage(pixels.SubImage(image.Rect(0, 0, 16, 1))))
			index := bank*45 + i
			positions[index] = motion.FrameParticle{X: float64(position&255)*2 + float64(control&1),
				Y:         float64(position>>8) + float64(control>>2&1)*256,
				VelocityX: float64(3-bank) * 2, Image: index}
		}
	}
	field, err := sprites.NewAnimatedField(sprites.AnimatedFieldConfig{
		Motion: motion.FrameFieldConfig{Count: len(positions), Spawn: func(i int, _ bool) motion.FrameParticle { return positions[i] },
			WrapX: &motion.FrameAxisWrap{Boundary: 512, Shift: 512, Inclusive: true}},
		Frames: images, ImageByParticle: true, OffsetX: -113, OffsetY: -34,
		Filter: ebiten.FilterNearest,
	})
	return field, images, err
}
