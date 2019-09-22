package utils

import (
	"encoding/base64"
	"github.com/google/uuid"
	"math"
	"math/rand"
	"strconv"
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
	str := id[:2]
	num, _ := strconv.ParseInt(str, 16, 32)
	num++
	return strconv.FormatInt(num, 16) + id[2:]
}

func CreateGuid() string {
	return uuid.New().String()
}

func RandomNo() int {
	return int(math.Floor(9001799254740992 * rand.Float64()))
}
