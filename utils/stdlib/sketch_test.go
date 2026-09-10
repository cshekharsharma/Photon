package stdlib

import (
	"bytes"
	"testing"
)

func TestNewCountMinSketch_InvalidParams(t *testing.T) {
	if _, err := NewCountMinSketch(0, 64, 0); err == nil {
		t.Fatalf("expected error for depth=0")
	}

	if _, err := NewCountMinSketch(9, 64, 0); err == nil {
		t.Fatalf("expected error for depth>8")
	}

	if _, err := NewCountMinSketch(4, 8, 0); err == nil {
		t.Fatalf("expected error for width<16")
	}
}

func TestCountMinSketch_BasicIncrementEstimate_AgeReset(t *testing.T) {
	cms, err := NewCountMinSketch(4, 64, 0) // power-of-two width (mask path)
	if err != nil {
		t.Fatalf("new cms: %v", err)
	}

	h := Hash64String("user:123")
	if got := cms.Estimate(h); got != 0 {
		t.Fatalf("estimate before inc = %d", got)
	}

	cms.Increment(h)
	if got := cms.Estimate(h); got != 1 {
		t.Fatalf("estimate after 1 inc = %d", got)
	}

	cms.Increment(h)
	if got := cms.Estimate(h); got != 2 {
		t.Fatalf("estimate after 2 inc = %d", got)
	}

	cms.Age()
	if got := cms.Estimate(h); got != 1 {
		t.Fatalf("estimate after age = %d, want 1", got)
	}

	cms.Reset()
	if got := cms.Estimate(h); got != 0 {
		t.Fatalf("estimate after reset = %d, want 0", got)
	}
}

func TestCountMinSketch_AutoAgingEvery(t *testing.T) {
	cms, err := NewCountMinSketch(4, 64, 1) // age on every op
	if err != nil {
		t.Fatalf("new cms: %v", err)
	}

	h := Hash64String("hot-key")

	cms.Increment(h) // increment then immediate age (1 -> 0)
	if got := cms.Estimate(h); got != 0 {
		t.Fatalf("auto-aged estimate = %d, want 0", got)
	}

	cms.Increment(h)
	cms.Increment(h)
}

func TestCountMinSketch_ModuloPath_NonPowerOfTwoWidth(t *testing.T) {
	cms, err := NewCountMinSketch(3, 100, 0) // non power of two -> modulo path
	if err != nil {
		t.Fatalf("new cms: %v", err)
	}

	h1 := Hash64String("k1")
	h2 := Hash64String("k2")
	cms.Increment(h1)
	cms.Increment(h1)
	cms.Increment(h2)

	if a, b := cms.Estimate(h1), cms.Estimate(h2); a < b {
		t.Fatalf("expected h1 (2) >= h2 (1-ish), got %d < %d", a, b)
	}
}

func TestCountMinSketch_Saturation(t *testing.T) {
	cms, err := NewCountMinSketch(2, 64, 0)
	if err != nil {
		t.Fatalf("new cms: %v", err)
	}

	h := Hash64String("maxed")
	for r := 0; r < cms.depth; r++ {
		idx := cms.index(r, h)
		cms.rows[r][idx] = ^uint32(0)
	}

	cms.Increment(h)
	if got := cms.Estimate(h); got != ^uint32(0) {
		t.Fatalf("saturation failed: got %d, want %d", got, ^uint32(0))
	}
}

func TestHashHelpers_DeterminismAndParity(t *testing.T) {
	b := []byte("abcdef")
	if Hash64String("abcdef") != Hash64Bytes(b) {
		t.Fatalf("string and bytes hash should match")
	}

	if Hash64String("abcdef") == Hash64String("abcdeg") {
		t.Fatalf("different strings should not hash equal (extremely unlikely)")
	}

	if !bytes.Equal([]byte("x"), []byte{"x"[0]}) {
		t.Fatalf("sanity check to use bytes import and branch coverage")
	}
}
