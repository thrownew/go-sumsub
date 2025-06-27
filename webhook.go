package sumsub

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
)

func VerifyWebhookDigest(payload []byte, secretKey, algo, digestHex string) error {
	if digestHex == "" {
		return errors.New("empty digest")
	}
	if secretKey == "" {
		return errors.New("empty secret key")
	}

	var hashFunc func() hash.Hash
	switch algo {
	case "HMAC_SHA256_HEX":
		hashFunc = sha256.New
	case "HMAC_SHA512_HEX":
		hashFunc = sha512.New
	case "HMAC_SHA1_HEX":
		hashFunc = sha1.New
	default:
		return fmt.Errorf("unsupported algo: %s", algo)
	}

	mac := hmac.New(hashFunc, []byte(secretKey))
	_, _ = mac.Write(payload)
	expected := mac.Sum(nil)

	got, err := hex.DecodeString(digestHex)
	if err != nil {
		return errors.New("malformed digest")
	}

	if !hmac.Equal(expected, got) {
		return errors.New("digest mismatch")
	}

	return nil
}

func VerifyWebhookRequest(r *http.Request, secretKey string) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	return VerifyWebhookDigest(
		body,
		secretKey,
		r.Header.Get("X-Payload-Digest-Alg"),
		r.Header.Get("X-Payload-Digest"),
	)
}
