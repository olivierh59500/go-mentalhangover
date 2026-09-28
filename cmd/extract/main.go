// Command extract reconstructs the reference disk's authored data transfers.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func main() {
	diskPath := flag.String("disk", "previous/Scoopex-MentalHangover.adf", "original disk image")
	modulePath := flag.String("module", "previous/madness.mod", "independent soundtrack reference")
	output := flag.String("out", "assets/raw", "decoded production data directory")
	flag.Parse()
	if err := run(*diskPath, *modulePath, *output); err != nil {
		log.Fatal(err)
	}
}

func run(diskPath, modulePath, output string) error {
	disk, err := os.ReadFile(diskPath)
	if err != nil {
		return err
	}
	if err := source.ValidateDisk(disk); err != nil {
		return err
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	type record struct {
		source.Region
		File, SHA256 string
		Bytes        int
	}
	var records []record
	for _, region := range source.Regions {
		data, err := source.DecodeRegion(disk, region)
		if err != nil {
			return err
		}
		file := region.Name + ".bin"
		if err := os.WriteFile(filepath.Join(output, file), data, 0644); err != nil {
			return err
		}
		records = append(records, record{region, file, fmt.Sprintf("%x", sha256.Sum256(data)), len(data)})
		if region.Name == "music" {
			module, err := source.ModuleBytes(data)
			if err != nil {
				return err
			}
			reference, err := os.ReadFile(modulePath)
			if err != nil {
				return err
			}
			if !bytes.Equal(module, reference) {
				return fmt.Errorf("extracted music differs from the supplied module")
			}
			if err := os.WriteFile(filepath.Join(output, "madness.mod"), module, 0644); err != nil {
				return err
			}
		}
	}
	resident, err := source.Resident(disk)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(output, "resident.bin"), resident, 0644); err != nil {
		return err
	}
	intro, err := source.DecodeRegion(disk, source.Regions[0])
	if err != nil {
		return err
	}
	eagle, err := source.Eagle(intro, resident)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(output, "eagle.png"))
	if err != nil {
		return err
	}
	encodeErr := png.Encode(file, eagle)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	manifest, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(output, "transfers.json"), append(manifest, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("Decoded %d cylinder transfers; original MOD matches the supplied reference exactly.\n", len(records))
	return nil
}
