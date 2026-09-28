package source

import (
	"encoding/binary"
	"fmt"
	"image"
)

type FilledSolid struct {
	Model   SolidModel
	Initial [6]int16
	Cues    []VectorCue
	Sines   []int16
	Palette *image.NRGBA
}

// FilledSolids follows the entry calls in the cylinder-38 segment. Its main
// patches the VBL comparison operand to one, so each source update is 50 Hz.
func FilledSolids(data []byte) ([]FilledSolid, error) {
	s := segment{data, 0x9000}
	quarter, err := s.read(0xb106, 1024)
	if err != nil {
		return nil, err
	}
	sines := make([]int16, 2048)
	for i := 0; i < 512; i++ {
		sines[i] = int16(binary.BigEndian.Uint16(quarter[i*2:]))
		sines[1023-i] = sines[i]
	}
	for i := 0; i < 1024; i++ {
		sines[i+1024] = -sines[i]
	}
	var solids []FilledSolid
	for _, definition := range []struct {
		name                         string
		points, edges, faces, motion uint32
		band                         int
	}{
		{"filled-cube", 0xac38, 0xac6a, 0xac84, 0xacaa, 10},
		{"faceted-solid", 0xacf0, 0xad6a, 0xadc0, 0xae46, 5}} {
		model, err := ReadSolidModel(data, 0x9000, definition.name, definition.points, definition.edges, definition.faces)
		if err != nil {
			return nil, err
		}
		initial, cues, err := ReadMotion(data, 0x9000, definition.motion)
		if err != nil {
			return nil, err
		}
		palette := image.NewNRGBA(image.Rect(0, 0, 54*3, 272))
		for phase := 0; phase < 54; phase++ {
			for y := 0; y < 272; y++ {
				for material := 0; material < 3; material++ {
					group := max(0, (y-1)/definition.band)
					if y >= 261 {
						palette.SetNRGBA(phase*3+material, y, RGB12(0))
						continue
					}
					word, err := s.word(0xae7e + uint32(material*216+((phase+group)%54)*2))
					if err != nil {
						return nil, err
					}
					palette.SetNRGBA(phase*3+material, y, RGB12(word))
				}
			}
		}
		if len(cues) == 0 {
			return nil, fmt.Errorf("source: missing filled-solid choreography")
		}
		solids = append(solids, FilledSolid{Model: model, Initial: initial, Cues: cues, Sines: sines, Palette: palette})
	}
	return solids, nil
}
