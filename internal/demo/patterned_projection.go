package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func patternedMatrix(angles [3]int16, sines []int16) [9]int16 {
	m, ok := motion.SampleWordEulerMatrix(wordMatrixRecipe(sines, true, true), angles)
	if !ok {
		panic("demo: invalid patterned matrix")
	}
	return m
}
func projectPatterned(points []source.Point3, matrix [9]int16, state [6]int16, output []source.Point2) error {
	return projectWordBank(len(points), func(i int) motion.WrappedPoint { p := points[i]; return motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z} }, matrix, [2]int16{state[3], state[4]}, state[5], wordProjection([2]int16{256, 100}, 7, 512), output)
}
