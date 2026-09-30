package demo

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
	"image/color"
)

// vectorEffect supplies the author choreography and depth color to DCK.
type vectorEffect struct {
	clock *vectorClock
	mesh  *effects.WordMesh
	color color.NRGBA
	ready bool
}

func newVectorEffect(model source.VectorModel, sines []int16) (*vectorEffect, error) {
	if len(sines) != 2048 || len(model.Cues) == 0 {
		return nil, fmt.Errorf("vector effect requires original clocks and waves")
	}
	e := &vectorEffect{clock: newVectorClock(model)}
	m := effects.WordMeshModel{Points: make([]motion.WrappedPoint, len(model.Points)), Faces: []effects.WordFace{{Contours: make([][]int, len(model.Contours))}}}
	for i, p := range model.Points {
		m.Points[i] = motion.WrappedPoint{X: p.X, Y: p.Y}
	}
	for i, c := range model.Contours {
		m.Faces[0].Contours[i] = c.Indices
	}
	var err error
	e.mesh, err = effects.NewWordMesh(effects.WordMeshConfig{Models: []effects.WordMeshModel{m}, Matrix: compiledWordMatrix(sines, false, false), Projection: wordProjection([2]int16{175, 136}, 8, 512), FillRule: ebiten.FillRuleEvenOdd, BatchTriangles: 512, Color: func(kit.Frame) color.NRGBA { return e.color }})
	return e, err
}
func (e *vectorEffect) Update(f kit.Frame) error {
	e.ready = e.clock.Step()
	if !e.ready {
		return nil
	}
	s := e.clock.state
	level := 16 - (max(2048, min(4096, int(s[5])))-2048)/128
	e.color = source.RGB12(uint16((2*level/16)<<8 | (4*level/16)<<4 | 15*level/16))
	if err := e.mesh.Update(f); err != nil {
		return err
	}
	return e.mesh.SetPose(effects.WordMeshPose{Angles: [3]int16{s[0], s[1], s[2]}, Translation: [2]int16{s[3], s[4]}, Depth: s[5] >> 1, Visible: true})
}
func (e *vectorEffect) Draw(dst *ebiten.Image) {
	if e.ready {
		e.mesh.Draw(dst)
	}
}
func (e *vectorEffect) Close() {
	if e.mesh != nil {
		e.mesh.Close()
	}
}
