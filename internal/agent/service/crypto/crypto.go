package crypto

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type CryptoManager struct {
	publicKey *rsa.PublicKey
	Enabled   bool
	Error     error
}

func NewCryptoManager(keyPath string) *CryptoManager {
	if keyPath == "" {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     errors.New("key Path not found"),
		}
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     fmt.Errorf("failed to read crypto key file %s: %w", keyPath, err),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     errors.New("failed to decode PEM block"),
		}
	}

	if block.Type != "PUBLIC KEY" && block.Type != "RSA PUBLIC KEY" {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     fmt.Errorf("invalid PEM type: %s", block.Type),
		}
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     fmt.Errorf("failed to parse public key: %w", err),
		}
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return &CryptoManager{
			publicKey: nil,
			Enabled:   false,
			Error:     errors.New("not RSA public key"),
		}
	}

	return &CryptoManager{
		publicKey: rsaKey,
		Enabled:   true,
		Error:     nil,
	}
}

// EncryptRSA encrypts data with a public key
func (cm *CryptoManager) EncryptRSA(data []byte) ([]byte, error) {
	if !cm.Enabled || cm.publicKey == nil {
		return data, nil // returning unencrypted data
	}

	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, nil, cm.publicKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}

	return ciphertext, nil
}
