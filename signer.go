package bitmex_request_signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Signer struct {
	keySecret string
}

func NewSigner(keySecret string) *Signer {
	return &Signer{
		keySecret: keySecret,
	}
}

func (s *Signer) Sign(data []byte) (signature string) {
	hasher := hmac.New(sha256.New, []byte(s.keySecret))
	hasher.Write(data)

	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *Signer) SignString(data string) (signature string) {
	return s.Sign([]byte(data))
}
