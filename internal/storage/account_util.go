package storage

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"time"
)

func ensureAccountSalt(in *Account) {
	if in == nil || in.Salt != "" {
		return
	}
	in.Salt = randomHex(32)
	if in.Salt == "" {
		in.Salt = "0000000000000000"
	}
}

func randomHex(n int) string {
	if n <= 0 {
		return ""
	}
	if n%2 != 0 {
		n++
	}
	b := make([]byte, n/2)
	if _, err := crand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(rand.Intn(256))
		}
	}
	return hex.EncodeToString(b)[:n]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
