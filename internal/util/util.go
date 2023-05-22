package util

import (
	"crypto/rand"
	"encoding/base64"
	mathrand "math/rand"
	"strconv"
	"strings"
)

func RandomBytes(nBytes int, encB64 bool) (s string, err error) {
	b := make([]byte, nBytes)
	_, err = rand.Read(b)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if encB64 {
		sb.WriteString(base64.URLEncoding.EncodeToString(b))
	} else {
		sb.Write(b)
	}

	return sb.String(), nil
}

func RandomSixDigitPin() string {
	var pin [6]string
	for i := 0; i < 6; i++ {
		pin[i] = strconv.Itoa(mathrand.Intn(10))
	}

	return strings.Join(pin[:], "")
}
