package demo

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/timeline"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// vectorClock binds authored durations/increments to the shared word program.
type vectorClock struct {
	state   [6]int16
	program *timeline.WordProgram
}

func newVectorClock(model source.VectorModel) *vectorClock {
	cues := make([]timeline.WordCue, len(model.Cues))
	for i, cue := range model.Cues {
		cues[i] = timeline.WordCue{Ticks: cue.Frames, Delta: cue.Delta[:]}
	}
	p, err := timeline.NewWordProgram(timeline.WordProgramConfig{Initial: model.Initial[:], Cues: cues})
	if err != nil {
		panic(err)
	}
	return &vectorClock{state: model.Initial, program: p}
}
func (c *vectorClock) Step() bool {
	if !c.program.Step() {
		return false
	}
	copy(c.state[:], c.program.State())
	return true
}

// vectorMatrix adapts the native fixture interface to DCK's signed matrix.
func vectorMatrix(state [6]int16, sines []int16) [6]int16 {
	m, ok := motion.SampleWordEulerMatrix(wordMatrixRecipe(sines, false, false), [3]int16{state[0], state[1], state[2]})
	if !ok {
		panic("demo: invalid vector matrix")
	}
	return [6]int16{m[0], m[1], m[2], m[3], m[4], m[5]}
}
func projectVector(points []source.Point2, matrix [6]int16, state [6]int16, output []source.Point2) error {
	m := [9]int16{matrix[0], matrix[1], matrix[2], matrix[3], matrix[4], matrix[5]}
	return projectWordBank(len(points), func(i int) motion.WrappedPoint { p := points[i]; return motion.WrappedPoint{X: p.X, Y: p.Y} }, m, [2]int16{state[3], state[4]}, state[5]>>1, wordProjection([2]int16{175, 136}, 8, 512), output)
}
