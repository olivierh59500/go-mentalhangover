// Package source decodes the original production's disk layout offline.
package source

import (
	"encoding/binary"
	"fmt"
)

const (
	DiskBytes     = 901120
	CylinderBytes = 22 * 512
	SectorKey     = uint32(0x21051972)
)

// Region is a transfer made by the resident routine at Amiga address 0x7ba.
// Its count is in cylinders: each iteration reads both eleven-sector tracks.
type Region struct {
	Name          string
	Cylinder      int
	Count         int
	MemoryAddress uint32
}

// Regions preserves the authored order, including the nonsequential cube load.
var Regions = []Region{
	{"intro", 3, 3, 0x20000},
	{"music", 6, 23, 0x40402},
	{"authors-ribbon", 29, 7, 0x9000},
	{"filled-vector", 36, 2, 0x9000},
	{"stencil", 38, 1, 0x9000},
	{"patterned-vectors", 41, 3, 0x8000},
	{"textured-cube", 39, 1, 0x9000},
	{"greetings", 44, 2, 0x9000},
	{"circle-scroll", 46, 2, 0x9000},
	{"final-reminder", 48, 1, 0x9000},
	{"checkerboard", 49, 11, 0x9000},
}

// ValidateDisk checks geometry and the end-around-carry Amiga boot checksum.
func ValidateDisk(data []byte) error {
	if len(data) != DiskBytes || string(data[:4]) != "DOS\x00" {
		return fmt.Errorf("source: expected the 880 KiB DOS0 reference disk")
	}
	var sum uint32
	for at := 0; at < 1024; at += 4 {
		word := binary.BigEndian.Uint32(data[at:])
		next := sum + word
		if next < sum {
			next++
		}
		sum = next
	}
	if sum != 0xffffffff {
		return fmt.Errorf("source: invalid boot checksum: %08x", sum)
	}
	return nil
}

// DecodeRegion reproduces the loader's post-MFM XOR, not an assumed file format.
// Source bytes stay untouched and no executable code is run by extraction.
func DecodeRegion(data []byte, region Region) ([]byte, error) {
	if region.Cylinder < 0 || region.Count < 1 || region.Cylinder >= len(data)/CylinderBytes ||
		region.Count > len(data)/CylinderBytes-region.Cylinder {
		return nil, fmt.Errorf("source: transfer %q is outside the disk", region.Name)
	}
	start, size := region.Cylinder*CylinderBytes, region.Count*CylinderBytes
	result := make([]byte, size)
	for offset := 0; offset < size; offset += 4 {
		word := binary.BigEndian.Uint32(data[start+offset:]) ^ SectorKey
		binary.BigEndian.PutUint32(result[offset:], word)
	}
	return result, nil
}

// Resident returns the relocated program copied from disk offset 0x41e to
// Amiga address 0x21e. The initial longword loop copies 0x1e59 longwords.
func Resident(data []byte) ([]byte, error) {
	const start, size = 0x41e, (0x1e58 + 1) * 4
	if len(data) < start+size || binary.BigEndian.Uint32(data[start:]) != 0x33fc7fff {
		return nil, fmt.Errorf("source: resident relocation differs from the reference")
	}
	return append([]byte(nil), data[start:start+size]...), nil
}

// ModuleBytes excludes the unrelated bytes following the original MOD samples.
func ModuleBytes(data []byte) ([]byte, error) {
	if len(data) < 1084 || string(data[1080:1084]) != "M.K." || data[950] < 1 || data[950] > 128 {
		return nil, fmt.Errorf("source: missing four-channel ProTracker module")
	}
	patterns := 0
	for _, pattern := range data[952 : 952+int(data[950])] {
		patterns = max(patterns, int(pattern)+1)
	}
	size := 1084 + patterns*1024
	for sample := 0; sample < 31; sample++ {
		size += int(binary.BigEndian.Uint16(data[20+sample*30+22:])) * 2
	}
	if size > len(data) {
		return nil, fmt.Errorf("source: truncated module samples")
	}
	return append([]byte(nil), data[:size]...), nil
}
