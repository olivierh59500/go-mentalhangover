// Command preview renders original decoded assets at the authored PAL cadence.
package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/go-mentalhangover/internal/demo"
)

func main() {
	page := flag.Bool("font", false, "preview the decoded original serif font")
	muted := flag.Bool("mute", false, "disable soundtrack playback")
	vector := flag.String("vector", "", "preview an original credit object: slayer, reward, uncle-tom or sign")
	title := flag.Bool("title", false, "preview the original five-plane Mental Hangover title")
	capturePath := flag.String("capture", "", "capture the original asset preview at four PAL checkpoints")
	flag.Parse()
	if *capturePath != "" {
		var game *demo.Preview
		err := capture.Run(capture.Config{Directory: *capturePath, Frames: []int{0, 50, 150, 300},
			Width: demo.Width, Height: demo.Height}, func() (ebiten.Game, error) {
			var err error
			game, err = demo.NewPreview(*page, true)
			if err == nil && *title {
				err = game.SelectTitle()
			}
			if err == nil && *vector != "" {
				err = game.SelectVector(*vector)
			}
			return game, err
		})
		if game != nil {
			game.Close()
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	game, err := demo.NewPreview(*page, *muted)
	if err != nil {
		log.Fatal(err)
	}
	defer game.Close()
	if *title {
		if err := game.SelectTitle(); err != nil {
			log.Fatal(err)
		}
	}
	if *vector != "" {
		if err := game.SelectVector(*vector); err != nil {
			log.Fatal(err)
		}
	}
	ebiten.SetTPS(demo.FPS)
	ebiten.SetWindowSize(demo.Width*3, demo.Height*3)
	ebiten.SetWindowTitle("Mental Hangover / Original asset preview")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
