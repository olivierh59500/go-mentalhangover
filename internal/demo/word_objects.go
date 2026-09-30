package demo

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func wordMatrixRecipe(sines []int16, solid, patterned bool) motion.WordEulerMatrixConfig {
	c := motion.WordEulerMatrixConfig{Sines: sines, Period: 4096, Quantum: 2, Quarter: 1024, ProductShift: 6, ThirdAxisShift: 7}
	if solid {
		c.MiddleAssociation = motion.WordMiddleXFirst
	}
	if patterned {
		c.Period, c.Quarter, c.ProductShift, c.ThirdAxisShift = 720, 180, 7, 8
	}
	return c
}
func compiledWordMatrix(sines []int16, solid, patterned bool) *motion.WordEulerMatrix {
	m, err := motion.NewWordEulerMatrix(wordMatrixRecipe(sines, solid, patterned))
	if err != nil {
		panic(err)
	}
	return m
}
func wordSolidModel(model source.SolidModel) effects.WordMeshModel {
	m := effects.WordMeshModel{Points: make([]motion.WrappedPoint, len(model.Points)), Faces: make([]effects.WordFace, len(model.Faces))}
	for i, p := range model.Points {
		m.Points[i] = motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z}
	}
	for i, f := range model.Faces {
		m.Faces[i] = effects.WordFace{Contours: [][]int{f.Vertices}, Material: int(f.Material)}
	}
	return m
}
func wordProjection(center [2]int16, shift uint8, bias int16) geometry.WordProjectionConfig {
	return geometry.WordProjectionConfig{Center: center, TranslationShift: shift, DepthShift: 9, DepthBias: bias}
}

// projectWordBank adapts decoded source structs for original CPU fixtures. Live
// effects use owned WordMesh banks, so no point import runs during animation.
func projectWordBank(count int, point func(int) motion.WrappedPoint, matrix [9]int16, translation [2]int16, depth int16, c geometry.WordProjectionConfig, output []source.Point2) error {
	if count != len(output) {
		return fmt.Errorf("demo: word projection output differs from geometry")
	}
	pose := geometry.WordPose{Matrix: matrix, Translation: translation, Depth: depth}
	for i := 0; i < count; i++ {
		p, err := geometry.ProjectWordPoint(point(i), pose, c)
		if err != nil {
			return err
		}
		output[i] = source.Point2{X: p.X, Y: p.Y}
	}
	return nil
}
