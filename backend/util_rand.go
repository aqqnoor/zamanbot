package main

import (
	"crypto/rand"
	"encoding/hex"
)

func RandString(n int) string {
	if n <= 0 {
		n = 16
	}
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
