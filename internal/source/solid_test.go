package source_test

import (
	"testing"

	"github.com/olivierh59500/go-mentalhangover/assets"
	"github.com/olivierh59500/go-mentalhangover/internal/source"
)

func TestBOBBanksRetainCompleteCubeAndPyramidTopology(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/filled-vector.bin")
	models, err := source.BOBModels(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || len(models[0].Points) != 8 || len(models[0].Edges) != 12 || len(models[0].Faces) != 6 ||
		len(models[1].Points) != 5 || len(models[1].Edges) != 8 || len(models[1].Faces) != 5 {
		t.Fatal("original BOB topology changed")
	}
	for _, model := range models {
		for _, face := range model.Faces {
			for i, a := range face.Vertices {
				b := face.Vertices[(i+1)%len(face.Vertices)]
				found := false
				for _, index := range face.Edges {
					edge := model.Edges[index]
					found = found || edge == [2]int{a, b} || edge == [2]int{b, a}
				}
				if !found {
					t.Fatal("source face has an invented boundary", model.Name, face)
				}
			}
		}
	}
}

func TestBOBControlsDoNotTreatBinaryZerosAsTextTerminators(t *testing.T) {
	data, _ := assets.Files.ReadFile("raw/filled-vector.bin")
	bobs, err := source.ReadBOBs(data)
	if err != nil {
		t.Fatal(err)
	}
	commands := 0
	for _, token := range bobs.Tokens {
		if token.Kind == 'b' {
			commands++
		}
	}
	if commands != 6 || bobs.Tokens[0].Kind != 'b' || bobs.Tokens[0].Values[0] != 0 || len(bobs.Text) < 400 {
		t.Fatal("packed BOB controls or their embedded zeros were lost")
	}
	if bobs.X[4] != 167 || bobs.X[5] != 169 || bobs.X[8] != 175 || bobs.X[9] != 177 {
		t.Fatal("word-aligned DMA address rounding changed horizontal motion", bobs.X[:10])
	}
	if _, err := source.ReadBOBs(data[:0x1000]); err == nil {
		t.Fatal("truncated BOB segment accepted")
	}
}
