// Command mentalhangover runs the reconstructed production units at PAL cadence.
package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/go-mentalhangover/internal/demo"
)

func main() {
	muted := flag.Bool("mute", false, "disable original module playback")
	directory := flag.String("capture", "", "write deterministic native screenshots")
	frames := flag.String("frames", "0,160,650,800,1120,1650,2400,3000,3800,4040,4800,5200,5800,6300,6700,7200,8100,8800,9500,10100,10500,11500,12100,12600,12900,13800,14800,15600,17000,17600", "capture ticks at 50 Hz")
	flag.Parse()
	if *directory != "" {
		var checkpoints []int
		for _, field := range strings.Split(*frames, ",") {
			tick, err := strconv.Atoi(field)
			if err != nil || tick < 0 || len(checkpoints) > 0 && tick <= checkpoints[len(checkpoints)-1] {
				log.Fatal("capture ticks must be nonnegative and increasing")
			}
			checkpoints = append(checkpoints, tick)
		}
		var game *demo.Opening
		err := capture.Run(capture.Config{Directory: *directory, Frames: checkpoints, Width: demo.Width, Height: demo.Height},
			func() (ebiten.Game, error) { var err error; game, err = demo.NewOpening(true); return game, err })
		if game != nil {
			game.Close()
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	game, err := demo.NewOpening(*muted)
	if err != nil {
		log.Fatal(err)
	}
	defer game.Close()
	ebiten.SetTPS(demo.FPS)
	ebiten.SetWindowSize(demo.Width*3, demo.Height*3)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Mental Hangover / Opening reconstruction")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
