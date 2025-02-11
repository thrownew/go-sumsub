package sumsub

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnitHMACSigner(t *testing.T) {
	signer := NewHMACSigner("api_secret")
	sign := signer.Sign(time.Unix(1712760187, 0), http.MethodPost, "/resources/accessTokens", []byte(`{"userId":"1000","levelName":"default","ttl":60}`))
	assert.Equal(t, "260c8893fb0317e8a714ce3bce9c16821649f1e27a49bd7d926fcf23942628c0", sign)
}
