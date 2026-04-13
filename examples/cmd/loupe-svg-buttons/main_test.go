package main

import (
	"testing"

	loupedeck "github.com/go-go-golems/loupedeck/pkg/device"
)

func TestResolveIconIndexesUsesRequestedOrder(t *testing.T) {
	lib := &loupedeck.SVGIconLibrary{
		Icons: []loupedeck.SVGIcon{
			{Name: "Finder"},
			{Name: "Trash"},
			{Name: "Clock"},
		},
	}
	idxs, err := resolveIconIndexes(lib, "clock,finder")
	if err != nil {
		t.Fatalf("resolveIconIndexes: %v", err)
	}
	if len(idxs) != 2 {
		t.Fatalf("len(idxs)=%d, want 2", len(idxs))
	}
	if idxs[0] != 2 || idxs[1] != 0 {
		t.Fatalf("unexpected indexes: %#v", idxs)
	}
}

func TestRotateIndexesAndBanking(t *testing.T) {
	idxs := rotateIndexes([]int{0, 1, 2, 3, 4}, 2)
	want := []int{2, 3, 4, 0, 1}
	for i := range want {
		if idxs[i] != want[i] {
			t.Fatalf("rotated[%d]=%d, want %d", i, idxs[i], want[i])
		}
	}
	if got, wantBanks := totalBanks(25, 12), 3; got != wantBanks {
		t.Fatalf("totalBanks=%d, want %d", got, wantBanks)
	}
}

func TestMakeBankPadsWithBlanks(t *testing.T) {
	prepared := []preparedIcon{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	icons := makeBank(prepared, 0)
	if len(icons) != len(grid) {
		t.Fatalf("len(icons)=%d, want %d", len(icons), len(grid))
	}
	if icons[0].Name != "A" || icons[1].Name != "B" || icons[2].Name != "C" {
		t.Fatalf("unexpected bank front: %#v %#v %#v", icons[0].Name, icons[1].Name, icons[2].Name)
	}
	for i := 3; i < len(icons); i++ {
		if icons[i].Name != "" {
			t.Fatalf("icons[%d].Name=%q, want blank", i, icons[i].Name)
		}
	}
}
