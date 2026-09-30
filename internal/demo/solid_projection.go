package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func solidMatrix(angles [3]int16, sines []int16) [9]int16 {
	m, ok := motion.SampleWordEulerMatrix(wordMatrixRecipe(sines, true, false), angles)
	if !ok {
		panic("demo: invalid solid matrix")
	}
	return m
}
func projectBOB(points []source.Point3, matrix [9]int16, depth int16, output []source.Point2) error {
	return projectWordBank(len(points), func(i int) motion.WrappedPoint { p := points[i]; return motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z} }, matrix, [2]int16{}, depth, wordProjection([2]int16{15, 14}, 0, 0), output)
}
func projectSolid(points []source.Point3, matrix [9]int16, state [6]int16, output []source.Point2) error {
	return projectWordBank(len(points), func(i int) motion.WrappedPoint { p := points[i]; return motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z} }, matrix, [2]int16{state[3], state[4]}, state[5]>>1, wordProjection([2]int16{351, 145}, 8, 512), output)
}
