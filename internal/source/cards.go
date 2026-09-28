package source

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TextCard preserves the original line count, vertical origin and null-terminated
// text. Empty lines are meaningful and still advance the blitter by 24 pixels.
type TextCard struct {
	Name    string
	Y       int
	Lines   []string
	Address uint32
}

func ReadTextCard(data []byte, base, address uint32, name string) (TextCard, error) {
	card := TextCard{Name: name, Address: address}
	if address < base || uint64(address-base)+4 > uint64(len(data)) {
		return card, fmt.Errorf("source: text descriptor %s is outside its segment", name)
	}
	p := int(address - base)
	count := int(binary.BigEndian.Uint16(data[p:]))
	card.Y = int(binary.BigEndian.Uint16(data[p+2:]))
	if count < 1 || count > 32 || card.Y > 272 {
		return card, fmt.Errorf("source: invalid text descriptor %s", name)
	}
	p += 4
	for i := 0; i < count; i++ {
		end := bytes.IndexByte(data[p:], 0)
		if end < 0 || end > 512 {
			return card, fmt.Errorf("source: unterminated text line in %s", name)
		}
		card.Lines = append(card.Lines, string(data[p:p+end]))
		p += end + 1
	}
	return card, nil
}

func AuthorCards(data []byte) ([]TextCard, error) {
	definitions := []struct {
		name    string
		address uint32
	}{
		{"presents", 0x9cb4}, {"new-demonstration", 0x9cc2}, {"called", 0x9cdc},
		{"coded-by", 0x9ce8}, {"graphics-by", 0x9cf6}, {"musics-by", 0x9d06},
		{"follow-sign", 0x9d14},
	}
	var cards []TextCard
	for _, definition := range definitions {
		card, err := ReadTextCard(data, 0x9000, definition.address, definition.name)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// ResidentCards are the full-screen interludes between loaded effect segments.
func ResidentCards(data []byte) ([]TextCard, error) {
	definitions := []struct {
		name    string
		address uint32
	}{
		{"scoopex-demo", 0x7ad6}, {"filled-vectors", 0x7b10},
		{"stencil-vectors", 0x7b3a}, {"always-remember", 0x7b66},
	}
	var cards []TextCard
	for _, definition := range definitions {
		card, err := ReadTextCard(data, 0x21e, definition.address, definition.name)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// PaletteMode identifies source copper operations, not generic alpha fades.
type PaletteMode uint8

const (
	PaletteTarget PaletteMode = iota
	PaletteGray
	PaletteFromWhite
	PaletteToBlack
	PaletteScaled128
	PaletteScaled32
	PaletteScaled64
)

func PaletteWord(word uint16, mode PaletteMode, level int) uint16 {
	var result uint16
	for shift := 0; shift <= 8; shift += 4 {
		value := int(word>>shift) & 15
		switch mode {
		case PaletteGray:
			value = max(0, min(15, level))
		case PaletteFromWhite:
			value = 15 - ((15 - value) * max(0, min(32, level)) >> 5)
		case PaletteToBlack:
			value = value * max(0, min(16, level)) >> 4
		case PaletteScaled128:
			value = value * max(0, min(127, level)) >> 7
		case PaletteScaled32:
			value = value * max(0, min(32, level)) >> 5
		case PaletteScaled64:
			value = value * max(0, min(64, level)) >> 6
		}
		result |= uint16(value) << shift
	}
	return result
}

// CardPalette follows the source's sixteen flash samples, thirty-three target
// samples, explicit hold and sixteen black samples. Tick is zero-based.
func CardPalette(tick, hold int) (PaletteMode, int) {
	if tick < 16 {
		return PaletteGray, max(0, tick)
	}
	if tick < 49 {
		return PaletteFromWhite, tick - 16
	}
	if tick < 49+hold {
		return PaletteTarget, 0
	}
	return PaletteToBlack, max(0, 15-(tick-49-hold))
}
