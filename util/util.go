package util

import (
	"math/rand"
	"time"
)

const s = "0123456789"

func GenerateVerifyCode(length int) string {
	code := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		code += string(s[rand.Intn(len(s))])
	}
	return code
}
