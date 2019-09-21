package utils

import (
	"encoding/base64"
	"github.com/google/uuid"
)

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

func GetNextNodeId(id string) string {
	r := []byte(id)
	r[1] += 1
	return string(r)
}

func CreateGuid() string {
	return uuid.New().String()
}
