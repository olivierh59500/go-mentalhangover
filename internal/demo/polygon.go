package demo

type polygonPoint struct{ x, y float64 }
type clippedFace struct {
	material uint8
	points   [10]polygonPoint
	count    int
}

// clipBOB confines the copied pixels to the original three-word, 32-row source
// tile. Projection can put edges outside that tile; DMA never copies them.
func clipBOB(points []polygonPoint) clippedFace {
	var result clippedFace
	copy(result.points[:], points)
	result.count = len(points)
	for axis := 0; axis < 4; axis++ {
		if result.count == 0 {
			break
		}
		var next [10]polygonPoint
		count := 0
		inside := func(p polygonPoint) bool {
			switch axis {
			case 0:
				return p.x >= 0
			case 1:
				return p.x <= 48
			case 2:
				return p.y >= 0
			default:
				return p.y <= 32
			}
		}
		intersection := func(a, b polygonPoint) polygonPoint {
			if axis < 2 {
				x := float64(0)
				if axis == 1 {
					x = 48
				}
				t := (x - a.x) / (b.x - a.x)
				return polygonPoint{x, a.y + t*(b.y-a.y)}
			}
			y := float64(0)
			if axis == 3 {
				y = 32
			}
			t := (y - a.y) / (b.y - a.y)
			return polygonPoint{a.x + t*(b.x-a.x), y}
		}
		a := result.points[result.count-1]
		for _, b := range result.points[:result.count] {
			if inside(a) != inside(b) {
				next[count] = intersection(a, b)
				count++
			}
			if inside(b) {
				next[count] = b
				count++
			}
			a = b
		}
		result.points, result.count = next, count
	}
	return result
}
