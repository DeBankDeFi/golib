package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
)

// Sha1 generate sha1 digest by the given string(s).
func Sha1(ss ...string) string {
	var buf bytes.Buffer
	for _, s := range ss {
		buf.WriteString(s)
	}
	h := sha1.New()
	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}
