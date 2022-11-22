package internal

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	gotils_strconv "github.com/savsgio/gotils/strconv"
)

const (
	HeaderSignature = "X-Signature-Ed25519"
	HeaderTimestamp = "X-Signature-Timestamp"
)

func verifySignature(ctx *gin.Context, publicKey ed25519.PublicKey, handler gin.HandlerFunc) {
	sig, ok := verifyEd25519Header(ctx.Request.Header.Get(HeaderSignature))
	if !ok {
		ctx.Status(http.StatusUnauthorized)

		return
	}

	timestamp := ctx.Request.Header.Get(HeaderTimestamp)

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)

		return
	}

	// Preserve original response.
	ctx.Request.Body.Close()
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	verified := ed25519.Verify(publicKey, append(gotils_strconv.S2B(timestamp), body...), sig)
	if !verified {
		ctx.String(http.StatusUnauthorized, ErrInvalidRequestSignature.Error())

		return
	}

	handler(ctx)
}

func verifyEd25519Header(value string) ([]byte, bool) {
	if value == "" {
		return nil, false
	}

	sig, err := hex.DecodeString(value)
	if err != nil {
		return nil, false
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return nil, false
	}

	return sig, true
}
