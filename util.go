package main

import (
	"math/rand/v2"
	"unsafe"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index (64 combinations)
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits (00111111 = 63)
)

func RandString6() string {
	// 1. Allocate exactly 6 bytes on the stack
	b := make([]byte, 6)

	// 2. Get 64 random bits out of a single call (fastest in math/rand/v2)
	cache := rand.Uint64()

	// 3. Unroll the loop manually to eliminate loop-counter overhead
	// Character 1
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[0] = letterBytes[idx]
	} else {
		b[0] = letterBytes[idx%len(letterBytes)]
	}
	cache >>= letterIdxBits
	// Character 2
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[1] = letterBytes[idx]
	} else {
		b[1] = letterBytes[idx%len(letterBytes)]
	}
	cache >>= letterIdxBits
	// Character 3
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[2] = letterBytes[idx]
	} else {
		b[2] = letterBytes[idx%len(letterBytes)]
	}
	cache >>= letterIdxBits
	// Character 4
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[3] = letterBytes[idx]
	} else {
		b[3] = letterBytes[idx%len(letterBytes)]
	}
	cache >>= letterIdxBits
	// Character 5
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[4] = letterBytes[idx]
	} else {
		b[4] = letterBytes[idx%len(letterBytes)]
	}
	cache >>= letterIdxBits
	// Character 6
	if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
		b[5] = letterBytes[idx]
	} else {
		b[5] = letterBytes[idx%len(letterBytes)]
	}

	// 4. Zero-allocation byte slice to string conversion
	return unsafe.String(&b[0], len(b))
}
