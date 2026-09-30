// Command checkframes fingerprints the full production across renderer changes.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/go-mentalhangover/internal/demo"
)

type sample struct {
	Scene  string `json:"scene"`
	Tick   int    `json:"tick"`
	SHA256 string `json:"sha256"`
}

type report struct {
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Rate    int      `json:"ticks_per_second"`
	Frames  int      `json:"drawn_frames"`
	Samples []sample `json:"samples"`
}

type probe struct {
	game         *demo.Game
	surface      *ebiten.Image
	pixels       []byte
	scene        string
	start, limit int
	report       report
	done         bool
	err          error
}

func (*probe) Layout(int, int) (int, int) { return demo.Width, demo.Height }

func (p *probe) Update() error {
	if p.done {
		return ebiten.Termination
	}
	return nil
}

func (p *probe) Draw(dst *ebiten.Image) {
	if p.done {
		return
	}
	if p.game == nil {
		p.game, p.err = demo.NewGame(true)
		if p.err != nil {
			p.done = true
			return
		}
		p.surface = render.NewSurface(demo.Width, demo.Height)
		p.pixels = make([]byte, demo.Width*demo.Height*4)
		p.report.Width, p.report.Height, p.report.Rate = demo.Width, demo.Height, demo.FPS
	}
	// Draw every source update, including the frames between fingerprint samples.
	for step := 0; step < 24 && !p.done; step++ {
		if p.report.Frames > 0 {
			if p.err = p.game.Update(); p.err != nil {
				p.done = true
				return
			}
		}
		name, tick := p.game.Position()
		if name != p.scene {
			p.scene, p.start = name, tick
		}
		p.game.Draw(p.surface)
		p.report.Frames++
		if tick-p.start < 12 || tick%113 == 0 || tick == p.limit {
			p.surface.ReadPixels(p.pixels)
			digest := sha256.Sum256(p.pixels)
			p.report.Samples = append(p.report.Samples, sample{Scene: name, Tick: tick, SHA256: hex.EncodeToString(digest[:])})
		}
		p.done = tick >= p.limit
	}
	dst.DrawImage(p.surface, nil)
}

func (p *probe) Close() {
	if p.game != nil {
		p.game.Close()
	}
	if p.surface != nil {
		p.surface.Deallocate()
	}
}

func main() {
	output := flag.String("output", "", "write complete-frame fingerprints as JSON")
	limit := flag.Int("last-tick", 24000, "last 50 Hz production tick to draw")
	flag.Parse()
	if *output == "" || *limit < 0 || *limit > 30000 {
		log.Fatal("-output and a last tick within 0..30000 are required")
	}
	p := &probe{limit: *limit}
	defer p.Close()
	ebiten.SetWindowSize(demo.Width, demo.Height)
	ebiten.SetWindowTitle("Mental Hangover / frame verification")
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(false)
	if err := ebiten.RunGame(p); err != nil {
		log.Fatal(err)
	}
	if p.err != nil {
		log.Fatal(p.err)
	}
	data, err := json.MarshalIndent(p.report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d full frames drawn; %d fingerprints recorded\n", p.report.Frames, len(p.report.Samples))
}
