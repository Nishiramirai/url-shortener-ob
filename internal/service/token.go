package service

import "crypto/rand"

const (
	alphabet    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	tokenLength = 10
)

func generateToken() (string, error) {
	// Небольшой запас под rejection sampling
	buffer := make([]byte, tokenLength+2)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	token := make([]byte, tokenLength)
	bufferIdx := 0

	for i := 0; i < tokenLength; i++ {
		for {
			if bufferIdx >= len(buffer) {
				var single [1]byte
				if _, err := rand.Read(single[:]); err != nil {
					return "", err
				}
				buffer = single[:]
				bufferIdx = 0
			}

			b := buffer[bufferIdx]
			bufferIdx++

			idx := int(b & 0x3F) // Маска 0x3F (0-63)
			if idx < len(alphabet) {
				token[i] = alphabet[idx]
				break
			}
		}
	}

	return string(token), nil
}
