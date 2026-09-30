package demo

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

type filledSolidEffect struct {
	clock   *vectorClock
	mesh    *effects.WordMesh
	palette *ebiten.Image
	phase   int
	ready   bool
}

func newFilledSolid(data source.FilledSolid) *filledSolidEffect {
	e := &filledSolidEffect{clock: newVectorClock(source.VectorModel{Initial: data.Initial, Cues: data.Cues}), palette: ebiten.NewImageFromImage(data.Palette)}
	var err error
	e.mesh, err = effects.NewWordMesh(effects.WordMeshConfig{Models: []effects.WordMeshModel{wordSolidModel(data.Model)}, Matrix: compiledWordMatrix(data.Sines, true, false), Projection: wordProjection([2]int16{351, 145}, 8, 512), Textures: []*ebiten.Image{e.palette}, CullBackFaces: true, Offset: geometry.Vec2{X: -176, Y: -1}, BatchTriangles: 256, UV: func(s effects.WordMaterialSample) geometry.Vec2 {
		return geometry.Vec2{X: float64(s.Phase*3) + float64(s.Material) - .5, Y: s.Screen.Y}
	}})
	if err != nil {
		e.Close()
		panic(err)
	}
	return e
}
func (e *filledSolidEffect) Update(f kit.Frame) error {
	e.ready = e.clock.Step()
	if !e.ready {
		return nil
	}
	e.phase = (e.phase + 1) % 54
	s := e.clock.state
	if err := e.mesh.Update(f); err != nil {
		return err
	}
	return e.mesh.SetPose(effects.WordMeshPose{Angles: [3]int16{s[0], s[1], s[2]}, Translation: [2]int16{s[3], s[4]}, Depth: s[5] >> 1, Visible: true, Phase: e.phase})
}
func (e *filledSolidEffect) Draw(dst *ebiten.Image) {
	if e.ready {
		e.mesh.Draw(dst)
	}
}
func (e *filledSolidEffect) Close() {
	if e.mesh != nil {
		e.mesh.Close()
	}
	if e.palette != nil {
		e.palette.Deallocate()
	}
}
