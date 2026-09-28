//go:build mental_vector_rendercheck

package demo

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestMain(m *testing.M) { os.Exit(testutil.RunGPU(m)) }

func TestVectorBatchUsesOneEvenOddFillForConcaveContoursAndHoles(t *testing.T) {
	// The outer L shape and inner rectangle share the source's single mask.
	model := source.VectorModel{Points: []source.Point2{{2, 2}, {22, 2}, {22, 8}, {10, 8}, {10, 22}, {2, 22},
		{4, 4}, {7, 4}, {7, 7}, {4, 7}},
		Contours: []source.Contour{{Material: 1, Indices: []int{0, 1, 2, 3, 4, 5, 0}},
			{Material: 1, Indices: []int{6, 7, 8, 9, 6}}},
		Cues: []source.VectorCue{{Frames: 1}}}
	effect, err := newVectorEffect(model, make([]int16, 2048))
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	copy(effect.points, model.Points)
	effect.color = color.NRGBA{B: 255, A: 255}
	effect.ready = true
	dst := ebiten.NewImage(24, 24)
	defer dst.Deallocate()
	effect.Draw(dst)
	pixels := image.NewRGBA(image.Rect(0, 0, 24, 24))
	dst.ReadPixels(pixels.Pix)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			outer := x >= 2 && x < 22 && y >= 2 && y < 8 || x >= 2 && x < 10 && y >= 8 && y < 22
			hole := x >= 4 && x < 7 && y >= 4 && y < 7
			want := outer && !hole
			got := pixels.RGBAAt(x, y).A != 0
			if got != want {
				t.Fatalf("parity mask differs at %d,%d: want %t, got %t", x, y, want, got)
			}
		}
	}
}

func TestCopperShaderMatchesAllOriginalNibbles(t *testing.T) {
	src, dst := ebiten.NewImage(64, 64), ebiten.NewImage(64, 64)
	defer src.Deallocate()
	defer dst.Deallocate()
	pixels := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for i := 1; i < 4096; i++ {
		pixels.SetNRGBA(i%64, i/64, source.RGB12(uint16(i)))
	}
	src.WritePixels(pixels.Pix)
	palette, err := newCopperPalette()
	if err != nil {
		t.Fatal(err)
	}
	defer palette.Close()
	for _, mode := range []source.PaletteMode{source.PaletteGray, source.PaletteFromWhite, source.PaletteToBlack, source.PaletteScaled128, source.PaletteScaled32, source.PaletteScaled64} {
		for _, level := range []int{0, 1, 7, 15, 16, 31, 32, 64, 127} {
			if (mode == source.PaletteGray || mode == source.PaletteToBlack) && level > 16 || (mode == source.PaletteFromWhite || mode == source.PaletteScaled32) && level > 32 || mode == source.PaletteScaled64 && level > 64 {
				continue
			}
			dst.Clear()
			palette.Draw(dst, src, mode, level)
			got := make([]byte, len(pixels.Pix))
			dst.ReadPixels(got)
			for i := 0; i < 4096; i++ {
				want := color.NRGBA{}
				if i != 0 {
					want = source.RGB12(source.PaletteWord(uint16(i), mode, level))
				}
				if !bytes.Equal(got[i*4:i*4+4], []byte{want.R, want.G, want.B, want.A}) {
					t.Fatalf("palette mode %d level %d color %03x: %v != %v", mode, level, i, got[i*4:i*4+4], want)
				}
			}
		}
	}
}

func TestOpeningRepeatedDrawPreservesClockAndVectorPose(t *testing.T) {
	game, err := NewGame(true)
	if err != nil {
		t.Fatal(err)
	}
	defer game.Close()
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for tick := 0; tick < 1700; tick++ {
		if tick > 0 {
			if err := game.Update(); err != nil {
				t.Fatal(err)
			}
		}
		if tick%137 != 0 {
			continue
		}
		clock, local := game.clock.Tick(), game.localTick
		state := game.vectors[0].clock.state
		game.Draw(dst)
		a := make([]byte, Width*Height*4)
		dst.ReadPixels(a)
		game.Draw(dst)
		b := make([]byte, len(a))
		dst.ReadPixels(b)
		if clock != game.clock.Tick() || local != game.localTick || state != game.vectors[0].clock.state || !bytes.Equal(a, b) {
			t.Fatal("drawing changes source timeline or pixels", tick)
		}
	}
}

func TestOpeningCardsMatchOriginalIntegerPensAndLineSteps(t *testing.T) {
	game, err := NewGame(true)
	if err != nil {
		t.Fatal(err)
	}
	defer game.Close()
	resident, _ := assets.Files.ReadFile("raw/resident.bin")
	authors, _ := assets.Files.ReadFile("raw/authors-ribbon.bin")
	font, advances, err := source.SerifFont(resident)
	if err != nil {
		t.Fatal(err)
	}
	cards, err := source.AuthorCards(authors)
	if err != nil {
		t.Fatal(err)
	}
	interludes, err := source.ResidentCards(resident)
	if err != nil {
		t.Fatal(err)
	}
	cards = append(cards, interludes...)
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for i, card := range cards {
		want := image.NewNRGBA(image.Rect(0, 0, Width, Height))
		for row, line := range card.Lines {
			width := 0
			for _, letter := range line {
				width += advances[int(letter)-32]
			}
			x, y := 176-width/2, card.Y+row*24
			for _, letter := range line {
				index := int(letter) - 32
				if letter != ' ' {
					draw.Draw(want, image.Rect(x, y, x+48, y+23), font, image.Pt(index*48, 0), draw.Over)
				}
				x += advances[index]
			}
		}
		dst.Clear()
		game.cards[i].Draw(dst)
		got := make([]byte, len(want.Pix))
		dst.ReadPixels(got)
		if !bytes.Equal(got, want.Pix) {
			t.Fatalf("card %s differs from source pen layout", card.Name)
		}
	}
}

func TestBOBFontMappingAndRepeatedDrawKeepClockAndPixels(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/filled-vector.bin")
	effect, err := newBOBEffect(data)
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for tick := 0; tick < 1350; tick++ {
		if err := effect.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		if tick%127 != 0 {
			continue
		}
		clock := effect.clock
		cursor, distance, angles := clock.cursor, clock.distance, clock.angles
		want := image.NewNRGBA(image.Rect(0, 0, Width, Height))
		letters := []rune(clock.data.Text)
		for i := clock.first; i < clock.fetched; i++ {
			if letters[i] == ' ' {
				continue
			}
			x := clock.origins[i] - distance
			draw.Draw(want, image.Rect(x, 15, x+16, 29), clock.data.Font, image.Pt((int(letters[i])-32)*16, 0), draw.Over)
		}
		dst.Clear()
		effect.Draw(dst)
		pixels := make([]byte, Width*Height*4)
		dst.ReadPixels(pixels)
		for y := 15; y < 29; y++ {
			if !bytes.Equal(pixels[y*Width*4:(y+1)*Width*4], want.Pix[y*Width*4:(y+1)*Width*4]) {
				t.Fatalf("BOB font pixels differ at update %d row %d", tick, y)
			}
		}
		dst.Clear()
		effect.Draw(dst)
		second := make([]byte, len(pixels))
		dst.ReadPixels(second)
		if cursor != clock.cursor || distance != clock.distance || angles != clock.angles || !bytes.Equal(pixels, second) {
			t.Fatal("BOB drawing advances the source clocks", tick)
		}
	}
}

func TestActualStencilEffectsHoldHalfRatePosesAndDrawWithoutMutation(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/patterned-vectors.bin")
	models, err := source.PatternedSolids(data)
	if err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for _, index := range []int{3, 6} {
		e := newPatterned(models[index])
		for i := 0; i < 80; i++ {
			if err := e.Update(kit.Frame{}); err != nil {
				t.Fatal(err)
			}
		}
		if err := e.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		pose := e.clock.state
		points := append([]source.Point2(nil), e.points...)
		dst.Clear()
		e.Draw(dst)
		a := make([]byte, Width*Height*4)
		dst.ReadPixels(a)
		if err := e.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		if e.clock.state != pose {
			t.Fatal("25 Hz stencil pose changed on an intervening PAL frame")
		}
		for j, p := range points {
			if e.points[j] != p {
				t.Fatal("held stencil projection changed")
			}
		}
		dst.Clear()
		e.Draw(dst)
		b := make([]byte, len(a))
		dst.ReadPixels(b)
		if !bytes.Equal(a, b) || e.clock.state != pose {
			t.Fatal("held stencil draw changed pixels or the pose")
		}
		e.Close()
	}
}

func TestGreetingDrawUsesCachedPagesAndDoesNotAdvanceSteering(t *testing.T) {
	data := starPagesForTest(t)
	palette, err := newCopperPalette()
	if err != nil {
		t.Fatal(err)
	}
	defer palette.Close()
	effect, err := newStarPages(data, palette)
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for tick := 0; tick < 1100; tick++ {
		effect.Update(kit.Frame{})
		if tick%97 != 0 {
			continue
		}
		angles, offset, frame := effect.clock.angles, effect.clock.offset, effect.clock.frame
		dst.Clear()
		effect.Draw(dst)
		a := make([]byte, Width*Height*4)
		dst.ReadPixels(a)
		dst.Clear()
		effect.Draw(dst)
		b := make([]byte, len(a))
		dst.ReadPixels(b)
		if angles != effect.clock.angles || offset != effect.clock.offset || frame != effect.clock.frame || !bytes.Equal(a, b) {
			t.Fatal("greeting draw changed steering or page pixels", tick)
		}
	}
}

func TestCircleOutlineDrawKeepsAuthoredProfilesAndMountainOcclusion(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/greetings.bin")
	curve, err := source.CirclePart(data)
	if err != nil {
		t.Fatal(err)
	}
	palette, err := newCopperPalette()
	if err != nil {
		t.Fatal(err)
	}
	defer palette.Close()
	effect, err := newCircle(curve, palette)
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	dst := ebiten.NewImage(Width, Height)
	defer dst.Deallocate()
	for tick := 0; tick < 2600; tick++ {
		effect.Update(kit.Frame{})
		if tick%193 != 0 {
			continue
		}
		clock := *effect.clock
		dst.Clear()
		effect.Draw(dst)
		a := make([]byte, Width*Height*4)
		dst.ReadPixels(a)
		for y := 0; y < 200; y++ {
			for x := 0; x < Width; x++ {
				if curve.Mountain.NRGBAAt(x, y).A == 0 {
					continue
				}
				pixel := a[((y+8)*Width+x)*4 : ((y+8)*Width+x)*4+4]
				if !bytes.Equal(pixel, []byte{0, 0, 0, 255}) {
					t.Fatal("foreground escaped the original mountain mask", tick, x, y)
				}
			}
		}
		dst.Clear()
		effect.Draw(dst)
		b := make([]byte, len(a))
		dst.ReadPixels(b)
		if effect.clock.angle != clock.angle || effect.clock.cursor != clock.cursor || effect.clock.rows != clock.rows || effect.clock.radiusPhase != clock.radiusPhase || !bytes.Equal(a, b) {
			t.Fatal("circle drawing changed the source profile or pixels", tick)
		}
	}
}
