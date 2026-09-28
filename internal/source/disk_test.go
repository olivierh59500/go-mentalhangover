package source

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestLoaderXORPreservesSectorOrderAndInput(t *testing.T) {
	disk := make([]byte, CylinderBytes*3)
	want := make([]byte, CylinderBytes)
	for at := 0; at < len(want); at += 4 {
		value := uint32(at*17 + 3)
		binary.BigEndian.PutUint32(want[at:], value)
		binary.BigEndian.PutUint32(disk[CylinderBytes+at:], value^SectorKey)
	}
	original := append([]byte(nil), disk...)
	got, err := DecodeRegion(disk, Region{Name: "fixture", Cylinder: 1, Count: 1})
	if err != nil || !bytes.Equal(got, want) || !bytes.Equal(disk, original) {
		t.Fatal("loader word order or input ownership changed", err)
	}
	for _, region := range []Region{{Cylinder: -1, Count: 1}, {Cylinder: 0, Count: 0}, {Cylinder: 2, Count: 2}, {Cylinder: 3, Count: 1}} {
		if _, err := DecodeRegion(disk, region); err == nil {
			t.Fatal("out-of-bounds transfer accepted", region)
		}
	}
}

func TestModuleExtentIncludesPatternsAndSamples(t *testing.T) {
	data := make([]byte, 1084+3*1024+10+64)
	copy(data[1080:], "M.K.")
	data[950], data[952], data[953] = 2, 0, 2
	binary.BigEndian.PutUint16(data[42:], 5)
	got, err := ModuleBytes(data)
	if err != nil || len(got) != len(data)-64 {
		t.Fatal("sample or highest-pattern extent changed", len(got), err)
	}
	if _, err := ModuleBytes(data[:len(got)-1]); err == nil {
		t.Fatal("truncated sample bank accepted")
	}
}
