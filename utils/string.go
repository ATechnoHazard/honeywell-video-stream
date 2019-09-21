package utils

import "encoding/base64"

func PadWithSpaces(pass string) string {
	p := ""
	for _, i := range pass {
		p = p + string(i) + string([]byte{0, 0, 0})
	}
	return p
}

func GenPass(clearText string) string {
	return base64.StdEncoding.EncodeToString([]byte(PadWithSpaces(clearText)))
}
