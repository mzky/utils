package common

import (
	"encoding/base64"
	"strings"
)

func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func Base64Decode(src string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(
		strings.TrimRight(StripUnprintable(src), "="))
}
