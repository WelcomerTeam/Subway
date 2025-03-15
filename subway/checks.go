package internal

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"

	gotils "github.com/savsgio/gotils/strconv"
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

	message := append(gotils.S2B(timestamp), body...)

	for _, key := range sub.publicKeys {
		if ed25519.Verify(key, message, sig) {
			return true
		}
	}

	return false
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
