package service

import "crypto/rand"

func generateToken() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	const tokenLength = 10

	bytes := make([]byte, tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	lenAlphabet := len(alphabet)
	for i, b := range bytes {
		bytes[i] = alphabet[b%byte(lenAlphabet)]
	}

	return string(bytes), nil
}
