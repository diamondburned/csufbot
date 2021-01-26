package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
)

// randToken generates a random token.
func randToken() (string, error) {
	buf := make([]byte, 16)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate randomness")
	}

	return base64.URLEncoding.EncodeToString(buf), nil
}
