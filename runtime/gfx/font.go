package gfx

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type FontOptions struct {
	Size    float64
	DPI     float64
	Index   int
	Hinting font.Hinting
}

type LoadedFont struct {
	path string
	opts FontOptions
	face font.Face
}

func (f *LoadedFont) Face() font.Face {
	if f == nil {
		return nil
	}
	return f.face
}

func (f *LoadedFont) Path() string {
	if f == nil {
		return ""
	}
	return f.path
}

func (f *LoadedFont) Options() FontOptions {
	if f == nil {
		return FontOptions{}
	}
	return f.opts
}

type fontCacheKey struct {
	path    string
	size    float64
	dpi     float64
	index   int
	hinting font.Hinting
}

var (
	fontCacheMu sync.Mutex
	fontCache   = map[fontCacheKey]*LoadedFont{}
)

func normalizeFontOptions(opts FontOptions) FontOptions {
	if opts.Size <= 0 {
		opts.Size = 12
	}
	if opts.DPI <= 0 {
		opts.DPI = 72
	}
	if opts.Index < 0 {
		opts.Index = 0
	}
	return opts
}

func LoadFont(path string, opts FontOptions) (*LoadedFont, error) {
	if path == "" {
		return nil, fmt.Errorf("gfx: font path must not be empty")
	}
	opts = normalizeFontOptions(opts)
	key := fontCacheKey{path: path, size: opts.Size, dpi: opts.DPI, index: opts.Index, hinting: opts.Hinting}

	fontCacheMu.Lock()
	if loaded := fontCache[key]; loaded != nil {
		fontCacheMu.Unlock()
		return loaded, nil
	}
	fontCacheMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("gfx: read font %q: %w", path, err)
	}

	face, err := parseFontFace(data, opts)
	if err != nil {
		return nil, err
	}
	loaded := &LoadedFont{path: path, opts: opts, face: face}

	fontCacheMu.Lock()
	defer fontCacheMu.Unlock()
	if existing := fontCache[key]; existing != nil {
		if closer, ok := face.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
		return existing, nil
	}
	fontCache[key] = loaded
	return loaded, nil
}

func parseFontFace(data []byte, opts FontOptions) (font.Face, error) {
	if collection, err := opentype.ParseCollection(data); err == nil {
		parsed, err := collection.Font(opts.Index)
		if err != nil {
			return nil, fmt.Errorf("gfx: font collection index %d: %w", opts.Index, err)
		}
		face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
			Size:    opts.Size,
			DPI:     opts.DPI,
			Hinting: opts.Hinting,
		})
		if err != nil {
			return nil, fmt.Errorf("gfx: create face from collection: %w", err)
		}
		return face, nil
	}

	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("gfx: parse font: %w", err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    opts.Size,
		DPI:     opts.DPI,
		Hinting: opts.Hinting,
	})
	if err != nil {
		return nil, fmt.Errorf("gfx: create face: %w", err)
	}
	return face, nil
}
