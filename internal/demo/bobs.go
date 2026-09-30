package demo

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// bobEffect submits the original projected copies through one retained DCK
// batch. A three-column source texture preserves copper colors by output row.
type bobEffect struct {
	clock              *bobClock
	models             []source.SolidModel
	fontImage, palette *ebiten.Image
	text               *scrolling.Scrolling
	mesh               *effects.WordMesh
	ready              bool
}

func newBOBEffect(data []byte) (*bobEffect, error) {
	decoded, err := source.ReadBOBs(data)
	if err != nil {
		return nil, err
	}
	models, err := source.BOBModels(data)
	if err != nil {
		return nil, err
	}
	effect := &bobEffect{clock: newBOBClock(decoded), models: models}
	effect.fontImage = ebiten.NewImageFromImage(decoded.Font)
	effect.palette = ebiten.NewImageFromImage(decoded.Palette)
	wordModels := make([]effects.WordMeshModel, len(models))
	for i, m := range models {
		wordModels[i] = wordSolidModel(m)
	}
	effect.mesh, err = effects.NewWordMesh(effects.WordMeshConfig{Models: wordModels, Matrix: compiledWordMatrix(decoded.Sines, true, false), Projection: wordProjection([2]int16{15, 14}, 0, 0), Textures: []*ebiten.Image{effect.palette}, CullBackFaces: true, FaceClip: &[4]float64{0, 0, 48, 32}, ClipVertexLimit: 10, BatchTriangles: 1024, MaxInstances: 256, Instance: func(index int, _ kit.Frame) effects.WordMeshInstance {
		x, y := effect.clock.Offset(index)
		return effects.WordMeshInstance{Offset: geometry.Vec2{X: float64(x), Y: float64(y)}}
	}, UV: func(s effects.WordMaterialSample) geometry.Vec2 {
		return geometry.Vec2{X: float64(s.Material) - .5, Y: s.Screen.Y}
	}})
	if err != nil {
		effect.Close()
		return nil, err
	}

	glyphs := make(map[rune]font.Glyph, 59)
	for i, advance := range decoded.Advances {
		glyphs[rune(i+32)] = font.Glyph{Rect: image.Rect(i*16, 0, i*16+16, 14), Advance: float64(advance)}
	}
	glyphs[' '] = font.Glyph{Advance: float64(decoded.Advances[0])}
	metrics, err := font.New(font.Config{Bounds: decoded.Font.Bounds(), Glyphs: glyphs, LineHeight: 14, SpaceAdvance: float64(decoded.Advances[0])})
	if err != nil {
		effect.Close()
		return nil, err
	}
	effect.text, err = scrolling.New(scrolling.Config{Insertion: &scrolling.InsertionConfig{Controller: effect.clock.program, ExternalClock: true, Y: 15, Fonts: map[string]scrolling.Face{"default": {Atlas: effect.fontImage, Metrics: metrics}}}})
	if err != nil {
		effect.Close()
		return nil, err
	}
	return effect, nil
}

func (e *bobEffect) Update(f kit.Frame) error {
	e.ready = e.clock.Step()
	if !e.ready || e.clock.depth == 4100 {
		return nil
	}
	if err := e.mesh.Update(f); err != nil {
		return err
	}
	if err := e.mesh.SetInstanceCount(e.clock.count); err != nil {
		return err
	}
	return e.mesh.SetPose(effects.WordMeshPose{Model: e.clock.model, Angles: e.clock.angles, Depth: e.clock.depth, Visible: true})
}

func (e *bobEffect) Draw(dst *ebiten.Image) {
	if !e.ready {
		return
	}
	c := e.clock
	e.text.Draw(dst)
	if c.depth == 4100 || c.count == 0 {
		return
	}
	e.mesh.Draw(dst)
}

func (e *bobEffect) Close() {
	if e.mesh != nil {
		e.mesh.Close()
	}
	if e.text != nil {
		e.text.Close()
	}
	if e.fontImage != nil {
		e.fontImage.Deallocate()
	}
	if e.palette != nil {
		e.palette.Deallocate()
	}
}
