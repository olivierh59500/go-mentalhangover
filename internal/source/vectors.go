package source

import (
	"encoding/binary"
	"fmt"
)

type Point2 struct{ X, Y int16 }

// Contour retains the repeated closing index and the authored face order.
// The original renderer XORs all contour edges before one even-odd fill.
type Contour struct {
	Material uint8
	Indices  []int
}

type VectorCue struct {
	Frames int
	Delta  [6]int16
}

type VectorModel struct {
	Name          string
	Points        []Point2
	Contours      []Contour
	Initial       [6]int16
	Cues          []VectorCue
	PointAddress  uint32
	FaceAddress   uint32
	MotionAddress uint32
}

// AuthorModels follows the resident call sites rather than guessing from shapes.
func AuthorModels(data []byte) ([]VectorModel, error) {
	definitions := []struct {
		name                   string
		points, faces, control uint32
	}{
		{"slayer", 0xac88, 0xad6e, 0xadc2},
		{"reward", 0xaab2, 0xabe4, 0xac50},
		{"uncle-tom", 0xadfa, 0xaebc, 0xaf0a},
		{"sign", 0xa972, 0xaa48, 0xaa88},
	}
	var models []VectorModel
	for _, definition := range definitions {
		model, err := ReadVectorModel(data, 0x9000, definition.name,
			definition.points, definition.faces, definition.control)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// ReadVectorModel decodes two-coordinate points, byte-packed closed polygons and
// six word-sized motion channels. Individual contours are not word-aligned.
func ReadVectorModel(data []byte, base uint32, name string, points, faces, control uint32) (VectorModel, error) {
	model := VectorModel{Name: name, PointAddress: points, FaceAddress: faces, MotionAddress: control}
	reader := func(address uint32, count int) ([]byte, error) {
		if address < base || uint64(address-base)+uint64(count) > uint64(len(data)) {
			return nil, fmt.Errorf("source: %s vector data outside segment", name)
		}
		return data[int(address-base) : int(address-base)+count], nil
	}
	header, err := reader(points, 2)
	if err != nil {
		return model, err
	}
	count := int(binary.BigEndian.Uint16(header))
	if count < 3 || count > 256 {
		return model, fmt.Errorf("source: invalid point count for %s", name)
	}
	bank, err := reader(points+2, count*4)
	if err != nil {
		return model, err
	}
	model.Points = make([]Point2, count)
	for i := range model.Points {
		model.Points[i] = Point2{int16(binary.BigEndian.Uint16(bank[i*4:])), int16(binary.BigEndian.Uint16(bank[i*4+2:]))}
	}
	header, err = reader(faces, 2)
	if err != nil {
		return model, err
	}
	faceCount := int(binary.BigEndian.Uint16(header))
	if faceCount < 1 || faceCount > 256 {
		return model, fmt.Errorf("source: invalid contour count for %s", name)
	}
	address := faces + 2
	for i := 0; i < faceCount; i++ {
		header, err := reader(address, 2)
		if err != nil {
			return model, err
		}
		vertexCount := int(header[1])
		if vertexCount < 3 {
			return model, fmt.Errorf("source: invalid contour size for %s", name)
		}
		indices, err := reader(address+2, vertexCount+1)
		if err != nil {
			return model, err
		}
		contour := Contour{Material: header[0], Indices: make([]int, len(indices))}
		for j, index := range indices {
			if int(index) >= count {
				return model, fmt.Errorf("source: %s contour references missing point", name)
			}
			contour.Indices[j] = int(index)
		}
		if contour.Indices[0] != contour.Indices[len(contour.Indices)-1] {
			return model, fmt.Errorf("source: %s contour is not closed", name)
		}
		model.Contours = append(model.Contours, contour)
		address += uint32(2 + vertexCount + 1)
	}
	bank, err = reader(control, 12)
	if err != nil {
		return model, err
	}
	for i := range model.Initial {
		model.Initial[i] = int16(binary.BigEndian.Uint16(bank[i*2:]))
	}
	address = control + 12
	for i := 0; i < 1024; i++ {
		header, err := reader(address, 2)
		if err != nil {
			return model, err
		}
		frames := int(binary.BigEndian.Uint16(header))
		if frames == 0x7fff {
			return model, nil
		}
		if frames < 1 || frames > 16384 {
			return model, fmt.Errorf("source: invalid vector cue for %s", name)
		}
		bank, err := reader(address+2, 12)
		if err != nil {
			return model, err
		}
		cue := VectorCue{Frames: frames}
		for channel := range cue.Delta {
			cue.Delta[channel] = int16(binary.BigEndian.Uint16(bank[channel*2:]))
		}
		model.Cues = append(model.Cues, cue)
		address += 14
	}
	return model, fmt.Errorf("source: vector cue sentinel missing for %s", name)
}

func AuthorSines(data []byte) ([]int16, error) {
	const offset, count = 0xaf50 - 0x9000, 2048
	if len(data) < offset+count*2 {
		return nil, fmt.Errorf("source: missing original vector sine table")
	}
	result := make([]int16, count)
	for i := range result {
		result[i] = int16(binary.BigEndian.Uint16(data[offset+i*2:]))
	}
	return result, nil
}
