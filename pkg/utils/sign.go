package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const AlgSignHeader = "HashSHA256"

func Sign(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func CheckSign(signature string, data *[]byte, key string) (bool, error) {
	expectedSignature, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(*data)
	return hmac.Equal(expectedSignature, h.Sum(nil)), nil
}
