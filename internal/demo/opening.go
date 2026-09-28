package demo

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sound"
	playback "github.com/olivierh59500/democonstructionkit/sound/ebiten"
	"github.com/olivierh59500/democonstructionkit/timeline"
	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

// Opening assembles the verified first production units. Later effects extend
// the director rather than playing the supplied reference movie.
type Opening struct {
	units        []openingUnit
	ranges       *timeline.CueRanges
	clock        *timeline.CueClock
	shared       *Preview
	palette      *copperPalette
	layer, title *ebiten.Image
	cards        []*scrolling.Scrolling
	vectors      []*vectorEffect
	unit         int
	localTick    int
	musicStarted bool
	muted        bool
	done         bool
	paletteMode  source.PaletteMode
	paletteLevel int
}

func NewOpening(muted bool) (*Opening, error) {
	data, err := assets.Files.ReadFile("raw/authors-ribbon.bin")
	if err != nil {
		return nil, err
	}
	models, err := source.AuthorModels(data)
	if err != nil {
		return nil, err
	}
	units, ranges, err := openingProgram(models)
	if err != nil {
		return nil, err
	}
	clock, err := timeline.NewCueClock(timeline.CueClockConfig{Rate: FPS})
	if err != nil {
		return nil, err
	}
	game := &Opening{units: units, ranges: ranges, clock: clock, unit: -1, muted: muted}
	game.shared, err = NewPreview(false, true)
	if err != nil {
		return nil, err
	}
	game.palette, err = newCopperPalette()
	if err != nil {
		game.Close()
		return nil, err
	}
	game.layer = render.NewSurface(Width, Height)
	pixels, err := source.MentalTitle(data)
	if err != nil {
		game.Close()
		return nil, err
	}
	game.title = ebiten.NewImageFromImage(pixels)
	resident, _ := assets.Files.ReadFile("raw/resident.bin")
	atlas, fontImage, err := serifAtlas(resident)
	if err != nil {
		game.Close()
		return nil, err
	}
	// Reuse the shared atlas surface; the preview caption does not own its image.
	game.shared.caption.Close()
	game.shared.fontImage.Deallocate()
	game.shared.fontImage = fontImage
	game.shared.caption = nil
	cardData, err := source.AuthorCards(data)
	if err != nil {
		game.Close()
		return nil, err
	}
	for _, card := range cardData {
		text, err := scrolling.New(scrolling.Config{Fonts: map[string]scrolling.Face{"default": atlas.Face()},
			Text: strings.Join(card.Lines, "\n"), Y: float64(card.Y),
			Map: func(sample scrolling.Sample, options *ebiten.DrawImageOptions) bool {
				// The original center divides a positive word before subtracting.
				options.GeoM.Translate(math.Ceil(sample.X)-sample.X, 0)
				return true
			},
			Page: &scrolling.PageConfig{Width: Width, LineHeight: 24, Align: scrolling.AlignCenter}})
		if err != nil {
			game.Close()
			return nil, err
		}
		game.cards = append(game.cards, text)
	}
	sines, err := source.AuthorSines(data)
	if err != nil {
		game.Close()
		return nil, err
	}
	for _, model := range models {
		vector, err := newVectorEffect(model, sines)
		if err != nil {
			game.Close()
			return nil, err
		}
		game.vectors = append(game.vectors, vector)
	}
	if err := game.prepare(); err != nil {
		game.Close()
		return nil, err
	}
	return game, nil
}

func (game *Opening) prepare() error {
	tick := game.clock.Tick()
	unit, _, active := game.ranges.At(float64(tick))
	if !active {
		game.done = true
		return nil
	}
	game.unit = unit
	data := game.units[unit]
	game.localTick = tick - data.Start
	game.paletteMode, game.paletteLevel = source.PaletteTarget, 0
	if data.Kind == "card" || data.Kind == "title" {
		game.paletteMode, game.paletteLevel = source.CardPalette(game.localTick, data.Hold)
	}
	if data.Kind == "eagle" {
		if tick >= openingMusicTick {
			game.paletteMode, game.paletteLevel = source.PaletteToBlack, 15-(tick-openingMusicTick)
		} else {
			game.paletteMode, game.paletteLevel = source.CardPalette(game.localTick, 100000)
		}
	}
	frame := kit.Frame{Tick: uint64(tick), Time: float64(tick) / FPS, Delta: 1.0 / FPS}
	if data.Kind == "vector" {
		if err := game.vectors[data.Model].Update(frame); err != nil {
			return err
		}
	}
	if data.Kind == "card" {
		return game.cards[data.Card].Update(frame)
	}
	return nil
}

func (game *Opening) Update() error {
	if game.done {
		return ebiten.Termination
	}
	game.clock.Step()
	frame := kit.Frame{Tick: uint64(game.clock.Tick()), Time: float64(game.clock.Tick()) / FPS, Delta: 1.0 / FPS}
	if err := game.shared.stars.Update(frame); err != nil {
		return err
	}
	if !game.musicStarted && game.clock.Tick() >= openingMusicTick {
		game.musicStarted = true
		if !game.muted {
			data, err := assets.Files.ReadFile("raw/madness.mod")
			if err != nil {
				return err
			}
			game.shared.player, err = playback.Open(nil, "madness.mod", data, sound.Options{Loop: true, Interpolation: true})
			if err != nil {
				return fmt.Errorf("start original module: %w", err)
			}
			game.shared.player.Play()
		}
	}
	return game.prepare()
}

func (game *Opening) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black)
	game.shared.stars.Draw(dst)
	if game.unit < 0 || game.done {
		return
	}
	data := game.units[game.unit]
	game.layer.Clear()
	switch data.Kind {
	case "eagle", "title":
		image, y := game.shared.eagle, 30.0
		if data.Kind == "title" {
			image, y = game.title, 62
		}
		options := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		options.GeoM.Translate(0, y)
		game.layer.DrawImage(image, &options)
	case "card":
		game.cards[data.Card].Draw(game.layer)
	case "vector":
		game.vectors[data.Model].Draw(game.layer)
	}
	game.palette.Draw(dst, game.layer, game.paletteMode, game.paletteLevel)
}

func (*Opening) Layout(int, int) (int, int) { return Width, Height }

func (game *Opening) Close() {
	for _, card := range game.cards {
		card.Close()
	}
	for _, vector := range game.vectors {
		vector.Close()
	}
	if game.layer != nil {
		game.layer.Deallocate()
	}
	if game.title != nil {
		game.title.Deallocate()
	}
	if game.palette != nil {
		game.palette.Close()
	}
	if game.shared != nil {
		game.shared.Close()
	}
}
