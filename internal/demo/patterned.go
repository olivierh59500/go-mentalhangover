package demo

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
	"image"
)

type patternedEffect struct {
	data     source.PatternedSolid
	clock    *vectorClock
	textures []*ebiten.Image
	mesh     *effects.WordMesh
	frame    int
	ready    bool
}

func newPatterned(data source.PatternedSolid) *patternedEffect {
	e := &patternedEffect{data: data, clock: newVectorClock(source.VectorModel{Initial: data.Initial, Cues: data.Cues})}
	m := effects.WordMeshModel{Points: make([]motion.WrappedPoint, len(data.Points)), Faces: make([]effects.WordFace, len(data.Faces))}
	for i, p := range data.Points {
		m.Points[i] = motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z}
	}
	for i, face := range data.Faces {
		m.Faces[i] = effects.WordFace{Contours: face.Contours, Material: int(face.Material), Texture: int(face.Material) - 1}
	}
	for _, texture := range data.Textures {
		e.textures = append(e.textures, ebiten.NewImageFromImage(texture))
	}
	e.data.Textures = nil
	var err error
	e.mesh, err = effects.NewWordMesh(effects.WordMeshConfig{Models: []effects.WordMeshModel{m}, Matrix: compiledWordMatrix(data.Sines, true, true), Projection: wordProjection([2]int16{256, 100}, 7, 512), Textures: e.textures, CullBackFaces: true, PerFaceBatch: true, FillRule: ebiten.FillRuleEvenOdd, Address: ebiten.AddressRepeat, BatchTriangles: 256, Offset: geometry.Vec2{X: -80, Y: 27}, DestinationClip: image.Rect(0, 30, Width, 227), UV: func(s effects.WordMaterialSample) geometry.Vec2 {
		return geometry.Vec2{X: s.Point.X - s.Center.X + 128, Y: s.Point.Y - s.Center.Y + 128}
	}})
	if err != nil {
		e.Close()
		panic(err)
	}
	return e
}
func (e *patternedEffect) Update(f kit.Frame) error {
	advance := e.frame%e.data.Period == 0
	e.frame++
	if !advance {
		return nil
	}
	e.ready = e.clock.Step()
	if !e.ready {
		return nil
	}
	for i := 0; i < 3; i++ {
		a := e.clock.state[i]
		if a > 718 {
			a -= 720
		} else if a < 0 {
			a += 720
		}
		e.clock.state[i] = a
	}
	s := e.clock.state
	if err := e.mesh.Update(f); err != nil {
		return err
	}
	return e.mesh.SetPose(effects.WordMeshPose{Angles: [3]int16{s[0], s[1], s[2]}, Translation: [2]int16{s[3], s[4]}, Depth: s[5], Visible: true})
}
func (e *patternedEffect) Level() int {
	depth := max(548, min(2590, int(e.clock.state[5])))
	return 127 - ((depth-548)&0xff0)/16
}
func (e *patternedEffect) Draw(dst *ebiten.Image) {
	if e.ready {
		e.mesh.Draw(dst)
	}
}
func (e *patternedEffect) Close() {
	if e.mesh != nil {
		e.mesh.Close()
	}
	for _, texture := range e.textures {
		texture.Deallocate()
	}
}
