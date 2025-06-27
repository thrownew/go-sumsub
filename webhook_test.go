package sumsub

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyWebhookDigest(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secretKey string
		algo      string
		digestHex string
		wantErr   error
	}{
		{
			name:      "Valid HMAC_SHA1_HEX",
			payload:   []byte("someText"),
			secretKey: "SoMe_SeCrEt_KeY",
			algo:      "HMAC_SHA1_HEX",
			digestHex: "f6e92ffe371718694d46e28436f76589312df8db",
		},
		{
			name:      "Valid HMAC_SHA256_HEX",
			payload:   []byte("test payload"),
			secretKey: "secret123",
			algo:      "HMAC_SHA256_HEX",
			digestHex: func() string {
				mac := hmac.New(sha256.New, []byte("secret123"))
				mac.Write([]byte("test payload"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Valid HMAC_SHA512_HEX",
			payload:   []byte("another payload"),
			secretKey: "another_secret",
			algo:      "HMAC_SHA512_HEX",
			digestHex: func() string {
				mac := hmac.New(sha512.New, []byte("another_secret"))
				mac.Write([]byte("another payload"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Empty digest",
			payload:   []byte("test"),
			secretKey: "secret",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "",
			wantErr:   errors.New("empty digest"),
		},
		{
			name:      "Empty secret key",
			payload:   []byte("test"),
			secretKey: "",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "abc123",
			wantErr:   errors.New("empty secret key"),
		},
		{
			name:      "Unsupported algorithm",
			payload:   []byte("test"),
			secretKey: "secret",
			algo:      "HMAC_MD5_HEX",
			digestHex: "abc123",
			wantErr:   errors.New("unsupported algo: HMAC_MD5_HEX"),
		},
		{
			name:      "Malformed digest hex",
			payload:   []byte("test"),
			secretKey: "secret",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "invalid_hex",
			wantErr:   fmt.Errorf("malformed digest"),
		},
		{
			name:      "Digest mismatch",
			payload:   []byte("test"),
			secretKey: "secret",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:   fmt.Errorf("digest mismatch"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyWebhookDigest(tt.payload, tt.secretKey, tt.algo, tt.digestHex)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyWebhookRequest(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		secretKey string
		algo      string
		digestHex string
		wantErr   error
	}{
		{
			name:      "Valid request with SHA256",
			body:      "test body content 1",
			secretKey: "secret_key_1",
			algo:      "HMAC_SHA256_HEX",
			digestHex: func() string {
				mac := hmac.New(sha256.New, []byte("secret_key_1"))
				mac.Write([]byte("test body content 1"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Valid request with SHA1",
			body:      "test body content 2",
			secretKey: "secret_key_2",
			algo:      "HMAC_SHA1_HEX",
			digestHex: func() string {
				mac := hmac.New(sha1.New, []byte("secret_key_2"))
				mac.Write([]byte("test body content 2"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Valid request with SHA512",
			body:      "test body content 3",
			secretKey: "secret_key_3",
			algo:      "HMAC_SHA512_HEX",
			digestHex: func() string {
				mac := hmac.New(sha512.New, []byte("secret_key_3"))
				mac.Write([]byte("test body content 3"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Empty body",
			body:      "",
			secretKey: "secret_key_4",
			algo:      "HMAC_SHA256_HEX",
			digestHex: func() string {
				mac := hmac.New(sha256.New, []byte("secret_key_4"))
				mac.Write([]byte(""))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
		},
		{
			name:      "Missing digest header",
			body:      "test",
			secretKey: "secret",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "",
			wantErr:   errors.New("empty digest"),
		},
		{
			name:      "Missing algorithm header",
			body:      "test",
			secretKey: "secret",
			algo:      "",
			digestHex: func() string {
				mac := hmac.New(sha256.New, []byte("secret"))
				mac.Write([]byte("test"))
				return hex.EncodeToString(mac.Sum(nil))
			}(),
			wantErr: errors.New("unsupported algo: "),
		},
		{
			name:      "Invalid digest format",
			body:      "test",
			secretKey: "secret",
			algo:      "HMAC_SHA256_HEX",
			digestHex: "invalid_hex",
			wantErr:   errors.New("malformed digest"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/webhook", strings.NewReader(tt.body))
			assert.Nil(t, err)

			req.Header.Set("X-Payload-Digest-Alg", tt.algo)
			req.Header.Set("X-Payload-Digest", tt.digestHex)

			err = VerifyWebhookRequest(req, tt.secretKey)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
