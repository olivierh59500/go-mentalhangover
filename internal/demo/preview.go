// Package demo contains the production's native composition and source adapters.
package demo

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sound"
	playback "github.com/olivierh59500/democonstructionkit/sound/ebiten"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

const Width, Height, FPS = 352, 272, 50

// Preview provides isolated asset views alongside the complete director.
type Preview struct {
	eagle, fontImage *ebiten.Image
	stars            *sprites.AnimatedField
	starImages       []*ebiten.Image
	caption          *scrolling.Scrolling
	player           *playback.Player
	frame            uint64
	page             bool
	vector           *vectorEffect
	title            bool
}

func NewPreview(page, muted bool) (*Preview, error) {
	resident, err := assets.Files.ReadFile("raw/resident.bin")
	if err != nil {
		return nil, err
	}
	intro, err := assets.Files.ReadFile("raw/intro.bin")
	if err != nil {
		return nil, err
	}
	pixels, err := source.Eagle(intro, resident)
	if err != nil {
		return nil, err
	}
	preview := &Preview{eagle: ebiten.NewImageFromImage(pixels), page: page}
	preview.stars, preview.starImages, err = originalStars(resident)
	if err != nil {
		preview.Close()
		return nil, err
	}
	atlas, image, err := serifAtlas(resident)
	if err != nil {
		preview.Close()
		return nil, err
	}
	preview.fontImage = image
	preview.caption, err = scrolling.New(scrolling.Config{
		Fonts: map[string]scrolling.Face{"default": atlas.Face()},
		Text:  "THIS IS A\nSCOOPEX DEMO", Y: 110,
		Page: &scrolling.PageConfig{Width: Width, LineHeight: 32, Align: scrolling.AlignCenter},
	})
	if err != nil {
		preview.Close()
		return nil, err
	}
	if !muted {
		data, err := assets.Files.ReadFile("raw/madness.mod")
		if err != nil {
			preview.Close()
			return nil, err
		}
		preview.player, err = playback.Open(nil, "madness.mod", data, sound.Options{Loop: true, Interpolation: true})
		if err != nil {
			preview.Close()
			return nil, fmt.Errorf("open original soundtrack: %w", err)
		}
		preview.player.Play()
	}
	return preview, nil
}

func (p *Preview) SelectTitle() error {
	data, err := assets.Files.ReadFile("raw/authors-ribbon.bin")
	if err != nil {
		return err
	}
	pixels, err := source.MentalTitle(data)
	if err != nil {
		return err
	}
	p.eagle.Deallocate()
	p.eagle = ebiten.NewImageFromImage(pixels)
	p.title = true
	return nil
}

// SelectVector replaces the diagnostic image with an original credit object.
func (p *Preview) SelectVector(name string) error {
	data, err := assets.Files.ReadFile("raw/authors-ribbon.bin")
	if err != nil {
		return err
	}
	models, err := source.AuthorModels(data)
	if err != nil {
		return err
	}
	sines, err := source.AuthorSines(data)
	if err != nil {
		return err
	}
	for _, model := range models {
		if model.Name != name {
			continue
		}
		vector, err := newVectorEffect(model, sines)
		if err != nil {
			return err
		}
		if p.vector != nil {
			p.vector.Close()
		}
		p.vector = vector
		return nil
	}
	return fmt.Errorf("unknown original credit object %q", name)
}

func (p *Preview) Update() error {
	p.frame++
	frame := kit.Frame{Tick: p.frame, Time: float64(p.frame) / FPS, Delta: 1.0 / FPS}
	if err := p.stars.Update(frame); err != nil {
		return err
	}
	if p.vector != nil {
		return p.vector.Update(frame)
	}
	return p.caption.Update(frame)
}

func (p *Preview) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black)
	p.stars.Draw(dst)
	if p.vector != nil {
		p.vector.Draw(dst)
		return
	}
	if p.page {
		p.caption.Draw(dst)
		return
	}
	var options ebiten.DrawImageOptions
	options.Filter = ebiten.FilterNearest
	y := 30.0
	if p.title {
		y = 62
	}
	options.GeoM.Translate(0, y)
	dst.DrawImage(p.eagle, &options)
}

func (*Preview) Layout(int, int) (int, int) { return Width, Height }

func (p *Preview) Close() {
	if p.vector != nil {
		p.vector.Close()
	}
	if p.player != nil {
		p.player.Close()
	}
	if p.caption != nil {
		p.caption.Close()
	}
	for _, image := range p.starImages {
		image.Deallocate()
	}
	if p.eagle != nil {
		p.eagle.Deallocate()
	}
	if p.fontImage != nil {
		p.fontImage.Deallocate()
	}
}
