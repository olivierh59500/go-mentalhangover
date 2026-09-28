package source

import (
	"encoding/binary"
	"fmt"
)

type Point3 struct{ X, Y, Z int16 }
type SolidFace struct {
	Material uint8
	Edges    []int
	Vertices []int
}
type SolidModel struct {
	Name   string
	Points []Point3
	Edges  [][2]int
	Faces  []SolidFace
}

type segment struct {
	data []byte
	base uint32
}

func (s segment) read(address uint32, count int) ([]byte, error) {
	if count < 0 || address < s.base || uint64(address-s.base)+uint64(count) > uint64(len(s.data)) {
		return nil, fmt.Errorf("source: address %x exceeds segment %x", address, s.base)
	}
	return s.data[int(address-s.base) : int(address-s.base)+count], nil
}
func (s segment) word(address uint32) (uint16, error) {
	data, err := s.read(address, 2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(data), nil
}

// ReadSolidModel decodes the edge-indexed face format used by the filled BOBs
// and later vector parts. Face boundaries are rebuilt from authored edges,
// retaining the first edge's orientation for the original visibility test.
func ReadSolidModel(data []byte, base uint32, name string, points, edges, faces uint32) (SolidModel, error) {
	model := SolidModel{Name: name}
	s := segment{data, base}
	n, err := s.word(points)
	if err != nil {
		return model, err
	}
	if n < 3 || n > 256 {
		return model, fmt.Errorf("source: invalid %s solid point count", name)
	}
	bank, err := s.read(points+2, int(n)*6)
	if err != nil {
		return model, err
	}
	for i := 0; i < int(n); i++ {
		model.Points = append(model.Points, Point3{int16(binary.BigEndian.Uint16(bank[i*6:])),
			int16(binary.BigEndian.Uint16(bank[i*6+2:])), int16(binary.BigEndian.Uint16(bank[i*6+4:]))})
	}
	n, err = s.word(edges)
	if err != nil {
		return model, err
	}
	if n < 3 || n > 512 {
		return model, fmt.Errorf("source: invalid %s edge count", name)
	}
	bank, err = s.read(edges+2, int(n)*2)
	if err != nil {
		return model, err
	}
	for i := 0; i < int(n); i++ {
		a, b := int(bank[i*2]), int(bank[i*2+1])
		if a >= len(model.Points) || b >= len(model.Points) || a == b {
			return model, fmt.Errorf("source: invalid %s solid edge", name)
		}
		model.Edges = append(model.Edges, [2]int{a, b})
	}
	n, err = s.word(faces)
	if err != nil {
		return model, err
	}
	if n < 1 || n > 256 {
		return model, fmt.Errorf("source: invalid %s face count", name)
	}
	address := faces + 2
	for i := 0; i < int(n); i++ {
		header, err := s.read(address, 2)
		if err != nil {
			return model, err
		}
		count := int(header[1])
		if count < 3 || count > len(model.Edges) {
			return model, fmt.Errorf("source: invalid %s face edge count", name)
		}
		indices, err := s.read(address+2, count)
		if err != nil {
			return model, err
		}
		face := SolidFace{Material: header[0]}
		for _, index := range indices {
			if int(index) >= len(model.Edges) {
				return model, fmt.Errorf("source: missing %s face edge", name)
			}
			face.Edges = append(face.Edges, int(index))
		}
		first := model.Edges[face.Edges[0]]
		second := model.Edges[face.Edges[1]]
		// The first pair is oriented exactly as the source's shared-corner test.
		if first[1] != second[0] && first[1] != second[1] {
			first[0], first[1] = first[1], first[0]
		}
		face.Vertices = []int{first[0], first[1]}
		used := make([]bool, count)
		used[0] = true
		for len(face.Vertices) < count {
			last := face.Vertices[len(face.Vertices)-1]
			found := false
			for j, index := range face.Edges {
				if used[j] {
					continue
				}
				edge := model.Edges[index]
				next := -1
				if edge[0] == last {
					next = edge[1]
				} else if edge[1] == last {
					next = edge[0]
				}
				if next < 0 || next == face.Vertices[0] {
					continue
				}
				face.Vertices = append(face.Vertices, next)
				used[j] = true
				found = true
				break
			}
			if !found {
				return model, fmt.Errorf("source: disconnected %s face", name)
			}
		}
		last := face.Vertices[len(face.Vertices)-1]
		closed := false
		for j, index := range face.Edges {
			if !used[j] {
				edge := model.Edges[index]
				closed = edge == [2]int{last, first[0]} || edge == [2]int{first[0], last}
			}
		}
		if !closed {
			return model, fmt.Errorf("source: open %s face", name)
		}
		model.Faces = append(model.Faces, face)
		address += uint32(count + 2)
	}
	return model, nil
}

func BOBModels(data []byte) ([]SolidModel, error) {
	var models []SolidModel
	for _, definition := range []struct {
		name                 string
		points, edges, faces uint32
	}{
		{"cube", 0xab30, 0xab62, 0xab7c}, {"pyramid", 0xaba2, 0xabc2, 0xabd4}} {
		model, err := ReadSolidModel(data, 0x9000, definition.name, definition.points, definition.edges, definition.faces)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
