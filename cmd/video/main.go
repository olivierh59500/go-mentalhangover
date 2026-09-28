// Command video exports the complete native sequence and its own module audio.
package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/video"
	"github.com/olivierh59500/go-mentalhangover/internal/demo"
)

type recordingHost struct{ *demo.Game }

// RecordingChapter supplies source-unit boundaries to DCK's video exporter.
func (h recordingHost) RecordingChapter() string {
	name, _ := h.Position()
	return strings.ReplaceAll(name, "-", " ")
}

func main() {
	config := video.Config{Output: "recordings/mental-hangover.mp4", Title: "Mental Hangover Go",
		Width: demo.Width * 2, Height: demo.Height * 2, FPS: demo.FPS, TPS: demo.FPS, SampleRate: 48000,
		Duration: 8 * time.Minute, PosterAt: 390 * time.Second}
	config.Flags(flag.CommandLine)
	flag.Parse()
	// The original ending loops indefinitely, so an export needs a finite limit.
	if config.Duration <= 0 {
		log.Fatal("Mental Hangover's ending loops; choose a positive -duration")
	}
	if err := video.Run(config, func() (ebiten.Game, error) {
		game, err := demo.NewGame(false)
		if err != nil {
			return nil, err
		}
		return recordingHost{game}, nil
	}); err != nil {
		log.Fatal(err)
	}
}
