package demo

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// newPointPlane supplies original point words, camera coefficients and depth
// colors. DCK owns projection, sparse raster preparation and batched rendering.
func newPointPlane(points []source.Point3, offset func() [3]int16, camera motion.WrappedPointProjectionConfig,
	collision sprites.PointPlaneCollision, origin image.Point, drawZero bool, depthMask func(uint16) byte, white *ebiten.Image) (*sprites.IndexedPointPlane, error) {
	projection, err := motion.NewWrappedPointProjection(camera)
	if err != nil {
		return nil, err
	}
	return sprites.NewIndexedPointPlane(sprites.IndexedPointPlaneConfig{Width: Width, Height: camera.Bounds.Max.Y,
		Count: len(points), Collision: collision, Offset: origin, DrawZero: drawZero, White: white,
		Palette: make([]color.NRGBA, 4), Sample: func(index int) (sprites.PointPlaneSample, bool) {
			p := points[index]
			pose, visible := projection.Project(motion.WrappedPoint{X: p.X, Y: p.Y, Z: p.Z}, offset())
			return sprites.PointPlaneSample{X: pose.X, Y: pose.Y, Mask: depthMask(pose.Depth)}, visible
		}})
}

func newStarPointPlane(points []source.Point3, offset func() [3]int16) (*sprites.IndexedPointPlane, error) {
	return newPointPlane(points, offset, motion.WrappedPointProjectionConfig{Mask: [3]uint16{511, 511, 2047}, Bias: [2]int16{-256, -256}, Numerator: 262144, DepthBias: 10, Shift: 9, Center: image.Pt(176, 143), Bounds: image.Rect(0, 0, Width, 286)},
		sprites.PointPlaneXOR, image.Pt(0, -10), false, func(depth uint16) byte {
			var mask byte
			if depth < 1700 {
				mask |= 1
			}
			if depth > 1000 {
				mask |= 2
			}
			return mask
		}, nil)
}

func newPerspectivePointPlane(points []source.Point3, offset func() [3]int16, white *ebiten.Image) (*sprites.IndexedPointPlane, error) {
	return newPointPlane(points, offset, motion.WrappedPointProjectionConfig{Mask: [3]uint16{1023, 511, 2047}, Bias: [2]int16{-512, -256}, Numerator: 263680, DepthBias: 390, Shift: 9, Center: image.Pt(176, 100), Bounds: image.Rect(0, 0, Width, 199)},
		sprites.PointPlaneOR, image.Pt(0, 8), true, func(depth uint16) byte {
			if depth < 1000 {
				return 1
			}
			if depth >= 1700 {
				return 3
			}
			return 2
		}, white)
}
