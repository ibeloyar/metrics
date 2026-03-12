package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type CryptoManager struct {
	privateKey *rsa.PrivateKey
	Enabled    bool
	Error      error
}

func NewCryptoManager(keyPath string) *CryptoManager {
	if keyPath == "" {
		return &CryptoManager{
			privateKey: nil,
			Enabled:    false,
			Error:      errors.New("key Path not found"),
		}
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return &CryptoManager{
			privateKey: nil,
			Enabled:    false,
			Error:      fmt.Errorf("read private key file %s error: %w", keyPath, err),
		}
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return &CryptoManager{
			privateKey: nil,
			Enabled:    false,
			Error:      errors.New("failed to decode PEM block"),
		}
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY" {
		return &CryptoManager{
			privateKey: nil,
			Enabled:    false,
			Error:      fmt.Errorf("invalid PEM type: %s", block.Type),
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return &CryptoManager{
			privateKey: nil,
			Enabled:    false,
			Error:      fmt.Errorf("failed to parse private key: %w", err),
		}
	}

	return &CryptoManager{
		privateKey: key,
		Enabled:    true,
		Error:      nil,
	}
}

// DecryptRSA encrypts data with a public key
func (cm *CryptoManager) DecryptRSA(data []byte) ([]byte, error) {
	if !cm.Enabled || cm.privateKey == nil {
		return data, nil // returning unencrypted data
	}

	hash := sha256.New()
	ciphertext, err := rsa.DecryptOAEP(hash, nil, cm.privateKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}

	return ciphertext, nil
}

func isEncrypted(r *http.Request) bool {
	return r.Header.Get("X-Crypto-Encrypted") == "true"
}

func decryptRequest(r *http.Request, privKey *rsa.PrivateKey) ([]byte, error) {
	gzReader, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, fmt.Errorf("gzip read failed: %w", err)
	}
	defer gzReader.Close()

	gzData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("gzip data read failed: %w", err)
	}

	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, nil, privKey, gzData, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa decrypt failed: %w", err)
	}

	return plaintext, nil
}

// CryptoMiddleware decrypts the request body if there is an X-Crypto-Encrypted header
func (cm *CryptoManager) CryptoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isEncrypted(r) || !cm.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		if cm == nil && cm.Enabled {
			http.Error(w, fmt.Sprintf("Crypto not configured: %s", cm.Error.Error()), http.StatusInternalServerError)
			return
		}

		decryptedBody, err := decryptRequest(r, cm.privateKey)
		if err != nil {
			log.Printf("decrypt failed: %v", err)
			http.Error(w, "Decryption failed", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
		r.Header.Set("Content-Length", fmt.Sprintf("%d", len(decryptedBody)))
		r.Header.Del("Content-Encoding") // убираем gzip после расшифровки

		next.ServeHTTP(w, r)
	})
}
