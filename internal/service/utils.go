package service

import "crypto/rand"

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

func generateAlias() (string) {
    b := make([]byte, 10)
	rand.Read(b)
    for i := range b {
        b[i] = alphabet[b[i]%63]
    }
    return string(b)
}