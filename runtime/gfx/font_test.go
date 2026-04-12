package gfx

import (
	"os"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func resetFontCacheForTest() {
	fontCacheMu.Lock()
	defer fontCacheMu.Unlock()
	fontCache = map[fontCacheKey]*LoadedFont{}
}

func TestLoadFontCachesRegularFontByKey(t *testing.T) {
	resetFontCacheForTest()
	tmp, err := os.CreateTemp(t.TempDir(), "font-*.ttf")
	if err != nil {
		t.Fatalf("create temp font: %v", err)
	}
	if _, err := tmp.Write(goregular.TTF); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp font: %v", err)
	}

	first, err := LoadFont(tmp.Name(), FontOptions{Size: 12, DPI: 72})
	if err != nil {
		t.Fatalf("load first font: %v", err)
	}
	second, err := LoadFont(tmp.Name(), FontOptions{Size: 12, DPI: 72})
	if err != nil {
		t.Fatalf("load second font: %v", err)
	}
	if first != second {
		t.Fatal("expected identical cache entry for same font key")
	}
	if first.Face() == nil {
		t.Fatal("expected loaded face")
	}
}

func TestLoadFontCreatesDistinctCacheEntriesForDifferentOptions(t *testing.T) {
	resetFontCacheForTest()
	tmp, err := os.CreateTemp(t.TempDir(), "font-*.ttf")
	if err != nil {
		t.Fatalf("create temp font: %v", err)
	}
	if _, err := tmp.Write(goregular.TTF); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp font: %v", err)
	}

	small, err := LoadFont(tmp.Name(), FontOptions{Size: 12, DPI: 72})
	if err != nil {
		t.Fatalf("load small font: %v", err)
	}
	large, err := LoadFont(tmp.Name(), FontOptions{Size: 18, DPI: 72})
	if err != nil {
		t.Fatalf("load large font: %v", err)
	}
	if small == large {
		t.Fatal("expected different cache entries for different font sizes")
	}
}

func TestLoadFontSupportsCollectionFileWhenAvailable(t *testing.T) {
	resetFontCacheForTest()
	const cjkCollection = "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"
	if _, err := os.Stat(cjkCollection); err != nil {
		t.Skipf("CJK collection not available: %v", err)
	}
	loaded, err := LoadFont(cjkCollection, FontOptions{Size: 14, DPI: 72, Index: 0})
	if err != nil {
		t.Fatalf("load collection font: %v", err)
	}
	if loaded == nil || loaded.Face() == nil {
		t.Fatal("expected collection-loaded font face")
	}
}
