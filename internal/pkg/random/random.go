package random

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

const alphaNumeric = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Hex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func AlphaNumeric(n int) (string, error) {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alphaNumeric)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphaNumeric[idx.Int64()]
	}
	return string(b), nil
}
