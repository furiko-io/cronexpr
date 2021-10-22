package cronexpr

import (
	"fmt"
	"sync/atomic"

	"github.com/cespare/xxhash/v2"
)

// WithHash returns a ParseOption that enables parsing of `H` symbols in a cron expression.
// The given hashID will be hashed using a deterministic hash function, which will be substituted
// where `H` is used in the cron expression.
func WithHash(hashID string) ParseOption {
	return &hashParseOption{hashID: hashID}
}

type hashParseOption struct {
	*baseOption
	hashID string
}

func (h *hashParseOption) GetPriority() int {
	// Parse WithHash before all other options.
	return 0
}

func (h *hashParseOption) Apply(expr *Expression) error {
	expr.hash = &hash{hashID: h.hashID}
	return nil
}

// WithHashEmptySeconds returns a ParseOption that will hash seconds by default if the seconds place is empty.
// Requires to be used in conjunction with WithHash, otherwise it will have no effect.
func WithHashEmptySeconds() ParseOption {
	return &hashSecondsParseOption{}
}

type hashSecondsParseOption struct {
	*baseOption
}

func (h *hashSecondsParseOption) Apply(expr *Expression) error {
	expr.hash.hashEmptySeconds = true
	return nil
}

// WithHashFields returns a ParseOption that will also hash the field name to make hashes less deterministic.
// For example, `H H * * * * *` will always hash the seconds and minutes to the same value, for example
// 00:37:37, 01:37:37, etc.
// Enabling this option will append additional keys to be hashed to introduce additional non-determinism.
func WithHashFields() ParseOption {
	return &hashFieldsParseOption{}
}

type hashFieldsParseOption struct {
	*baseOption
}

func (h *hashFieldsParseOption) Apply(expr *Expression) error {
	expr.hash.hashFields = true
	return nil
}

type hash struct {
	// ID to hash.
	hashID string

	// Whether we should hash an empty seconds field by default.
	// If set to true, `0 * * * * *` will be parsed as `H 0 * * * * *` internally.
	hashEmptySeconds bool

	// Whether we should also hash the field name (e.g. "day-of-week").
	// This helps to give more non-deterministic hashes for expressions with multiple `H` tokens
	// with intervals of the same size.
	hashFields bool

	// Memoized value that was previously computed.
	value  uint64
	hashed bool
}

// GetValueForField returns the value of the hash, given a specific field (e.g. "day-of-week").
// See GetValue for more information.
func (h *hash) GetValueForField(min, max int, field string) int {
	hash := h
	if h.hashFields {
		hash = hash.AddSuffix(field)
	}
	return hash.GetValue(min, max)
}

// GetValue returns the materialized value for the hash within the bounds of min and max (both inclusive).
// Because the hash value is an unsigned integer, the conversion of unsigned to signed integers may overflow.
func (h *hash) GetValue(min, max int) int {
	mod := max - min + 1
	v := int(h.getValue()) % mod // note: may overflow here
	if v < 0 {
		v += mod
	}
	return min + v
}

func (h *hash) getValue() uint64 {
	if !h.hashed {
		atomic.StoreUint64(&h.value, HashString(h.hashID))
		h.hashed = true
	}
	return atomic.LoadUint64(&h.value)
}

// AddSuffix returns a new hash with suffix added.
// This helps to make hashes more non-deterministic within a single cron expression with multiple H tokens.
func (h *hash) AddSuffix(suffix string) *hash {
	return &hash{hashID: h.hashID + ":" + suffix}
}

// HashString takes in a string, and deterministically hashes the string to return a 64-bit unsigned integer.
// This does not use a cryptographic hash function for both speed and simplicity.
// Specifically, we use xxHash which has excellent performance and anti-collision and avalanche properties,
// as outlined here: https://cyan4973.github.io/xxHash/
func HashString(str string) uint64 {
	return xxhash.Sum64String(str)
}

// makeErrorNoHashInput returns a new error for to display the string
// containing a H token when there is no hash input.
func makeErrorNoHashInput(s string) error {
	return fmt.Errorf("hash requested without using WithHash: %v", s)
}
