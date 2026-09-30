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
	"runtime"
	"sort"
	"time"

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
	Timing  *timing  `json:"cpu_submission_timing,omitempty"`
}

type timings struct {
	MeanUS float64 `json:"mean_us"`
	P95US  float64 `json:"p95_us"`
	MaxUS  float64 `json:"max_us"`
}

type timing struct {
	Update      timings `json:"update"`
	Draw        timings `json:"draw"`
	PeakHeapMiB float64 `json:"peak_go_heap_mib"`
}

func summarize(values []time.Duration) timings {
	if len(values) == 0 {
		return timings{}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	var total time.Duration
	for _, value := range values {
		total += value
	}
	return timings{MeanUS: float64(total) / float64(len(values)) / 1000,
		P95US: float64(values[(len(values)-1)*95/100]) / 1000,
		MaxUS: float64(values[len(values)-1]) / 1000}
}

type probe struct {
	game           *demo.Game
	surface        *ebiten.Image
	pixels         []byte
	scene          string
	start, limit   int
	report         report
	done           bool
	err            error
	measure        bool
	updates, draws []time.Duration
	peakHeap       uint64
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
		if p.measure {
			p.updates = make([]time.Duration, 0, p.limit+1)
			p.draws = make([]time.Duration, 0, p.limit+1)
		}
	}
	// Draw every source update, including the frames between fingerprint samples.
	for step := 0; step < 24 && !p.done; step++ {
		if p.report.Frames > 0 {
			var started time.Time
			if p.measure {
				started = time.Now()
			}
			if p.err = p.game.Update(); p.err != nil {
				p.done = true
				return
			}
			if p.measure {
				p.updates = append(p.updates, time.Since(started))
			}
		}
		name, tick := p.game.Position()
		if name != p.scene {
			p.scene, p.start = name, tick
		}
		var started time.Time
		if p.measure {
			started = time.Now()
		}
		p.game.Draw(p.surface)
		if p.measure {
			p.draws = append(p.draws, time.Since(started))
		}
		p.report.Frames++
		if tick-p.start < 12 || tick%113 == 0 || tick == p.limit {
			p.surface.ReadPixels(p.pixels)
			digest := sha256.Sum256(p.pixels)
			p.report.Samples = append(p.report.Samples, sample{Scene: name, Tick: tick, SHA256: hex.EncodeToString(digest[:])})
			if p.measure {
				var memory runtime.MemStats
				runtime.ReadMemStats(&memory)
				p.peakHeap = max(p.peakHeap, memory.HeapAlloc)
			}
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
	measure := flag.Bool("timing", false, "measure CPU update/draw submission and Go heap; excludes GPU completion and readback")
	flag.Parse()
	if *output == "" || *limit < 0 || *limit > 30000 {
		log.Fatal("-output and a last tick within 0..30000 are required")
	}
	p := &probe{limit: *limit, measure: *measure}
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
	if p.measure {
		p.report.Timing = &timing{Update: summarize(p.updates), Draw: summarize(p.draws), PeakHeapMiB: float64(p.peakHeap) / (1 << 20)}
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
