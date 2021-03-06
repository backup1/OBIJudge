package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
)

// Compress a []byte with gzip
func compress(data []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

// Decompress a gzipped []byte
func decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}

	b, _ := ioutil.ReadAll(r)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// Generate an alphanumeric key of specified size
func generateKey(size int) ([]byte, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1"

	key := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return []byte{}, err
	}

	for i, b := range key {
		key[i] = letters[b%byte(len(letters))]
	}

	return key, nil
}

// Encrypt encrypts data using 128-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func encrypt(plaintext []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 128-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func decrypt(ciphertext []byte, key []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// Strips all repeated space characters and newlines from a string
func strip(s string) string {
	out := strings.Builder{}
	out.Grow(len(s))

	white := false
	for _, c := range s {
		if unicode.IsSpace(c) {
			if !white {
				out.WriteString(" ")
			}
			white = true
		} else {
			out.WriteRune(c)
			white = false
		}
	}

	return out.String()
}
