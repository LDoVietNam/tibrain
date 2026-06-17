package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

// getMasterKey derives a 32-byte AES key from either TI_MASTER_KEY env var
// or a machine-unique combination of Hostname, User Home Dir, and a custom salt.
func getMasterKey() []byte {
	masterKeySource := os.Getenv("TI_MASTER_KEY")
	if masterKeySource == "" {
		// Fallback: machine-unique key derivation
		hostname, _ := os.Hostname()
		homeDir, _ := os.UserHomeDir()
		// Combine hostname + homeDir + static salt
		masterKeySource = fmt.Sprintf("%s:%s:TiEcosystemSalt2026", hostname, homeDir)
	}

	hash := sha256.Sum256([]byte(masterKeySource))
	return hash[:]
}

// EncryptToken encrypts plaintext using AES-GCM-256.
// Prepends with a special prefix "ENC:" followed by hex-encoded ciphertext.
func EncryptToken(plaintext []byte) (string, error) {
	key := getMasterKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return "ENC:" + hex.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts ciphertext hex prefixed with "ENC:".
// If not prefixed with "ENC:", returns the input as is (plaintext fallback).
func DecryptToken(encrypted string) ([]byte, error) {
	if len(encrypted) < 4 || encrypted[:4] != "ENC:" {
		// Plaintext fallback
		return []byte(encrypted), nil
	}

	ciphertextHex := encrypted[4:]
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %v", err)
	}

	key := getMasterKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}

	return plaintext, nil
}
