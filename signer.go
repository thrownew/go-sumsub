package sumsub

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"
)

type (
	HashSigner struct {
		mu   sync.Mutex
		hash hash.Hash
	}
)

func NewHMACSigner(secret string) *HashSigner {
	return NewHashSigner(hmac.New(sha256.New, []byte(secret)))
}

func NewHashSigner(h hash.Hash) *HashSigner {
	return &HashSigner{
		hash: h,
	}
}

func (s *HashSigner) Sign(t time.Time, method, uri string, payload []byte) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash.Reset()
	_, _ = s.hash.Write([]byte(fmt.Sprintf("%d%s%s%s", t.Unix(), method, uri, string(payload)))) //nolint:staticcheck
	return strings.ToLower(hex.EncodeToString(s.hash.Sum(nil)))
}
