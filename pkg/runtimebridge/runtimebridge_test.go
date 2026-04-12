package runtimebridge

import (
	"context"
	"testing"

	"github.com/dop251/goja"
)

func TestStoreLookupDelete(t *testing.T) {
	vm := goja.New()
	bindings := Bindings{
		Context: context.Background(),
		Values:  map[string]any{"name": "loupedeck"},
	}

	Store(vm, bindings)
	got, ok := Lookup(vm)
	if !ok {
		t.Fatal("expected lookup to succeed")
	}
	if got.Context == nil {
		t.Fatal("expected context to be stored")
	}
	if got.Values["name"] != "loupedeck" {
		t.Fatalf("unexpected stored value: %#v", got.Values)
	}

	Delete(vm)
	if _, ok := Lookup(vm); ok {
		t.Fatal("expected bindings to be deleted")
	}
}
