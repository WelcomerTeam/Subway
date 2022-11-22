package internal

import (
	"crypto/ed25519"
	"encoding/hex"
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
	signature := ctx.Request.Header.Get(HeaderSignature)

	if !verifyEd25519Header(signature) {
		ctx.Status(http.StatusUnauthorized)

		return
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		ctx.Status(http.StatusBadRequest)

		return
	}

	timestamp := ctx.Request.Header.Get(HeaderTimestamp)

	defer ctx.Request.Body.Close()

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)

		return
	}

	verified := ed25519.Verify(publicKey, append(gotils_strconv.S2B(timestamp), body...), sig)
	if verified {
		handler(ctx)

		return
	}

	ctx.String(http.StatusBadRequest, ErrInvalidRequestSignature.Error())
}

func verifyEd25519Header(value string) bool {
	if value == "" {
		return false
	}

	sig, err := hex.DecodeString(value)
	if err != nil {
		return false
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return false
	}

	return true
}
