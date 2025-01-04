package web

import "crypto/rand"

func getRandomData(length int) []byte {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		logError("Failed to generate random data: %s", err)
	}
	return data
}
