package ui

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"testing"

	static "github.com/hanzoai/status/web"
)

// brandsInBinary lists the brand directories actually shipped, so adding a brand
// puts it under test without anyone remembering to add it here.
func brandsInBinary(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(static.FileSystem, static.RootPath+brandRoot[:len(brandRoot)-1])
	if err != nil {
		t.Fatal(err)
	}
	var brands []string
	for _, entry := range entries {
		if entry.IsDir() {
			brands = append(brands, entry.Name())
		}
	}
	if len(brands) < 2 {
		t.Fatalf("expected the binary to carry several brands, found %v", brands)
	}
	return brands
}

// TestEveryBrandIsComplete proves each shipped brand can dress the whole page.
// A brand missing one size is how a page ends up falling back to whatever file
// happens to answer that path.
func TestEveryBrandIsComplete(t *testing.T) {
	for _, brand := range brandsInBinary(t) {
		t.Run(brand, func(t *testing.T) {
			if err := iconsFor(brand).validate(); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestNoBrandWearsAnothersMark is the defect this file exists for. One binary
// serves every brand we run. When brand-neutral copies of the artwork sat at the
// web root, every brand's /apple-touch-icon.png, /logo-192x192.png and
// /favicon.ico were the same bytes — Hanzo's — so saving status.lux.network to an
// iOS home screen saved the Hanzo mark. Distinct bytes per brand is the property
// that was silently false, so it is the property worth asserting.
func TestNoBrandWearsAnothersMark(t *testing.T) {
	brands := brandsInBinary(t)
	for _, slot := range []struct {
		name string
		of   func(Icons) string
	}{
		{"favicon.svg", func(i Icons) string { return i.SVG }},
		{"favicon.ico", func(i Icons) string { return i.ICO }},
		{"favicon-16.png", func(i Icons) string { return i.PNG16 }},
		{"favicon-32.png", func(i Icons) string { return i.PNG32 }},
		{"icon-180.png", func(i Icons) string { return i.Touch180 }},
		{"icon-192.png", func(i Icons) string { return i.PNG192 }},
		{"icon-512.png", func(i Icons) string { return i.PNG512 }},
		{"logo.svg", func(i Icons) string { return i.Logo }},
	} {
		t.Run(slot.name, func(t *testing.T) {
			owner := make(map[[32]byte]string, len(brands))
			for _, brand := range brands {
				asset := slot.of(iconsFor(brand))
				body, err := fs.ReadFile(static.FileSystem, static.RootPath+asset)
				if err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(body)
				if previous, taken := owner[sum]; taken {
					t.Errorf("%s and %s ship identical %s — one of them is wearing the other's mark", previous, brand, slot.name)
					continue
				}
				owner[sum] = brand
			}
		})
	}
}

// TestUnknownBrandFailsToBoot keeps a typo loud. Resolving to files that are not
// there would leave the page wearing nothing, and nothing looks enough like a
// slow load that it would ship.
func TestUnknownBrandFailsToBoot(t *testing.T) {
	cfg := &Config{Brand: "hanzoo"}
	if err := cfg.ValidateAndSetDefaults(); !errors.Is(err, ErrBrandAssetMissing) {
		t.Errorf("expected %v, got %v", ErrBrandAssetMissing, err)
	}
}
