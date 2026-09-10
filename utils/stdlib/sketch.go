// Package stdlib implements reusable, high‑performance primitives.
//
// This file provides a Count‑Min Sketch (CMS) used by TinyLFU
// admission policies to maintain approximate frequency counts under tight
// memory bounds.
//
// Design highlights
//   - Fixed memory: depth × width counters, independent of key cardinality
//   - 64‑bit FNV‑1a base hash + per‑row salts (no heap allocs per op)
//   - Saturating uint32 counters with periodic aging (halving)
//   - Zero external dependencies
//
// Recommended configuration
//
//	depth = 4 rows, width = power‑of‑two (e.g., 1<<16) is a good default.
//	agingEvery = N ops between global halving; tune to workload recency.
//
// Concurrency: the sketch is NOT internally synchronized. Callers should
// coordinate access (e.g., shard‑level locks). This mirrors the LRU core.
package stdlib

import (
	"fmt"
)

var minimumCountMinSketchWidth uint64 = 16
var maximumCountMinSketchDepth int = 8

// CountMinSketch maintains approximate per‑key frequency counts using a compact
// 2D array of counters addressed by multiple hash functions. It never
// underestimates frequency; collisions can overestimate counts.
type CountMinSketch struct {
	depth      int
	width      uint64
	mask       uint64     // width‑1 when width is power‑of‑two; used for fast modulo
	rows       [][]uint32 // counters[depth][width]
	salt       [8]uint64  // per‑row salts (supports up to 8 rows)
	n          uint64     // ops since last aging
	agingEvery uint64     // 0 disables automatic aging
}

// NewCountMinSketch constructs a sketch with the given depth (rows), width
// (columns per row), and optional automatic aging interval.
//
//   - depth must be 1..8
//   - width must be >= 16; power‑of‑two is recommended for speed
func NewCountMinSketch(depth int, width uint64, agingEvery uint64) (*CountMinSketch, error) {
	if depth <= 0 || depth > maximumCountMinSketchDepth {
		return nil, fmt.Errorf("countMinSketch: depth must be in [1,%d]", maximumCountMinSketchDepth)
	}

	if width < minimumCountMinSketchWidth {
		return nil, fmt.Errorf("countMinSketch: width must be >= %d", minimumCountMinSketchWidth)
	}

	rows := make([][]uint32, depth)
	for i := range rows {
		rows[i] = make([]uint32, width)
	}

	cms := &CountMinSketch{
		depth:      depth,
		width:      width,
		mask:       width - 1,
		rows:       rows,
		salt:       deriveSalts(),
		agingEvery: agingEvery,
	}

	// If width is not a power of two, disable fast mask and fallback to modulo.
	if width&(width-1) != 0 {
		cms.mask = 0
	}
	return cms, nil
}

// Increment records one observation for the provided 64‑bit key hash.
// The hash should be well‑distributed; see Hash64* helpers below.
func (c *CountMinSketch) Increment(h uint64) {
	for r := 0; r < c.depth; r++ {
		idx := c.index(r, h)
		if v := c.rows[r][idx]; v != ^uint32(0) {
			c.rows[r][idx] = v + 1 // saturating increment
		}
	}

	c.n++
	if c.agingEvery != 0 && c.n%c.agingEvery == 0 {
		c.Age()
	}
}

// Estimate returns the approximate frequency for the provided 64‑bit hash.
// The value is the minimum counter across rows, which guarantees no
// underestimation.
func (c *CountMinSketch) Estimate(h uint64) uint32 {
	min := ^uint32(0)

	for r := 0; r < c.depth; r++ {
		if v := c.rows[r][c.index(r, h)]; v < min {
			min = v
		}
	}

	return min
}

// Age halves all counters, biasing the sketch toward recent history. This is a
// bulk operation and should be called infrequently (e.g., every 50k–500k ops).
func (c *CountMinSketch) Age() {
	for r := 0; r < c.depth; r++ {
		row := c.rows[r]

		for i := uint64(0); i < c.width; i++ {
			row[i] >>= 1
		}
	}
}

// Reset zeroes all counters and the op counter. Useful for tests or hard resets.
func (c *CountMinSketch) Reset() {
	for r := 0; r < c.depth; r++ {
		for i := range c.rows[r] {
			c.rows[r][i] = 0
		}
	}

	c.n = 0
}

// index returns the bucket index for (row r, key hash h).
func (c *CountMinSketch) index(r int, h uint64) uint64 {
	// Derive a per‑row hash via XOR with a salt and a final mix.
	h2 := mix64(h ^ c.salt[r])

	if c.mask != 0 { // fast path for power‑of‑two widths
		return h2 & c.mask
	}

	return h2 % c.width
}

// ------------------------- Hash utilities -------------------------

// Hash64String computes a stable 64‑bit hash for strings using FNV‑1a.
func Hash64String(s string) uint64 {
	var h uint64 = 1469598103934665603

	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}

	return h
}

// Hash64Bytes computes FNV‑1a over a byte slice.
func Hash64Bytes(b []byte) uint64 {
	var h uint64 = 1469598103934665603

	for i := range b {
		h ^= uint64(b[i])
		h *= 1099511628211
	}

	return h
}

// mix64 is a fast 64‑bit bit‑mixer (SplitMix64 finalizer).
func mix64(z uint64) uint64 {
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// deriveSalts returns 8 distinct odd constants to decorrelate rows.
func deriveSalts() [8]uint64 {
	return [8]uint64{
		0x9e3779b97f4a7c15, // golden ratio 64‑bit
		0xc2b2ae3d27d4eb4f,
		0x165667b19e3779f9,
		0x85ebca77c2b2ae63,
		0x27d4eb2f165667c5,
		0x94d049bb133111eb,
		0xbf58476d1ce4e5b9,
		0x2545f4914f6cdd1d,
	}
}
