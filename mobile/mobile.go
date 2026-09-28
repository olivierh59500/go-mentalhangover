// Package mobile hosts the same PAL production as the desktop command.
package mobile

import (
	"log"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/olivierh59500/go-mentalhangover/internal/demo"
)

type host struct {
	game         *demo.Game
	requests     chan int
	lastUnit     string
	lastReport   time.Time
	verification bool
}

var gameHost = &host{requests: make(chan int, 1)}

func init() {
	ebiten.SetTPS(demo.FPS)
	enginemobile.SetGame(gameHost)
}

func (h *host) Update() error {
	if h.game == nil {
		tick := 0
		select {
		case tick = <-h.requests:
			h.verification = true
		default:
		}
		var err error
		h.game, err = demo.NewGame(h.verification && tick > 0)
		if err != nil {
			return err
		}
		if h.verification && tick > 0 {
			if err := h.game.FastForward(tick); err != nil {
				return err
			}
		}
	}
	if err := h.game.Update(); err != nil {
		return err
	}
	unit, tick := h.game.Position()
	if unit != h.lastUnit {
		log.Printf("mental_scene name=%s tick=%d", unit, tick)
		h.lastUnit = unit
	}
	if h.verification && time.Since(h.lastReport) >= 5*time.Second {
		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		log.Printf("mental_verify tick=%d tps=%.1f fps=%.1f heap_mib=%.1f", tick, ebiten.ActualTPS(), ebiten.ActualFPS(), float64(memory.HeapAlloc)/(1<<20))
		h.lastReport = time.Now()
	}
	return nil
}

func (h *host) Draw(dst *ebiten.Image) {
	if h.game != nil {
		h.game.Draw(dst)
	}
}

func (*host) Layout(int, int) (int, int) { return demo.Width, demo.Height }

// ConfigureVerification selects a checkpoint before the view starts.
// Tick zero keeps normal audio playback; a positive tick uses a muted seek.
// Scene clocks still advance at 50 Hz after the initial deterministic seek.
func ConfigureVerification(tick int) bool {
	if tick < 0 || tick > 30000 {
		return false
	}
	select {
	case gameHost.requests <- tick:
		return true
	default:
		return false
	}
}

// Dummy keeps the Go entry point visible to the mobile binding generator.
func Dummy() {}
