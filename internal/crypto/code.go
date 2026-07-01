package crypto

import (
	"crypto/rand"
	"fmt"
)

func GenerateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "ERROR-CODE-TEMP"
	}

	code := make([]byte, 16)
	for i, b := range bytes {
		code[i] = charset[b%byte(len(charset))]
	}

	return fmt.Sprintf("%s-%s-%s-%s",
		code[0:4], code[4:8], code[8:12], code[12:16])
}
