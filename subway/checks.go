package internal

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"

	gotils_strconv "github.com/savsgio/gotils/strconv"
)

const (
	HeaderSignature = "X-Signature-Ed25519"
	HeaderTimestamp = "X-Signature-Timestamp"
)

func (sub *Subway) verifySignature(request *http.Request, body []byte) bool {
	sig, ok := verifyEd25519Header(request.Header.Get(HeaderSignature))
	if !ok {
		return false
	}

	timestamp := request.Header.Get(HeaderTimestamp)

	return ed25519.Verify(sub.publicKey, append(gotils_strconv.S2B(timestamp), body...), sig)
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
