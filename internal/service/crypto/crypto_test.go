package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewCryptoManager(t *testing.T) {
	tests := []struct {
		name        string
		keyPath     string
		wantEnabled bool
		wantErrMsg  string
	}{
		{"empty", "", false, "key Path not found"},
		{"notfound", "nofile.key", false, "read private key file nofile.key error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCryptoManager(tt.keyPath)
			if cm.Enabled != tt.wantEnabled {
				t.Errorf("Enabled=%v want %v", cm.Enabled, tt.wantEnabled)
			}
			if tt.wantErrMsg != "" && (!strings.Contains(cm.Error.Error(), tt.wantErrMsg)) {
				t.Errorf("Error=%v want %q", cm.Error, tt.wantErrMsg)
			}
		})
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		header string
		want   bool
	}{
		{"true", true}, {"false", false}, {"", false},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Crypto-Encrypted", tt.header)
			if isEncrypted(req) != tt.want {
				t.Errorf("isEncrypted(%q)=%v want %v", tt.header, isEncrypted(req), tt.want)
			}
		})
	}
}

func TestCryptoMiddleware(t *testing.T) {
	cmDisabled := &CryptoManager{Enabled: false}
	cmNil := &CryptoManager{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// crypto disabled
	middleware1 := cmDisabled.CryptoMiddleware(next)
	req1, _ := http.NewRequest("POST", "/", nil)
	req1.Header.Set("X-Crypto-Encrypted", "true")
	rr1 := httptest.NewRecorder()
	middleware1.ServeHTTP(rr1, req1)

	// nil check
	middleware2 := cmNil.CryptoMiddleware(next)
	req2, _ := http.NewRequest("POST", "/", nil)
	req2.Header.Set("X-Crypto-Encrypted", "true")
	rr2 := httptest.NewRecorder()
	middleware2.ServeHTTP(rr2, req2)

	// without encrypted header
	req3, _ := http.NewRequest("POST", "/", nil)
	middleware1.ServeHTTP(rr2, req3)
}

func TestDecryptRSA(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		privKey *rsa.PrivateKey
		input   []byte
		wantErr bool
	}{
		{
			name:    "disabled",
			enabled: false,
			privKey: nil,
			input:   []byte("test"),
			wantErr: false,
		},
		{
			name:    "enabled_nil_key",
			enabled: true,
			privKey: nil,
			input:   []byte("test"),
			wantErr: false,
		},
		{
			name:    "enabled_valid_key",
			enabled: true,
			privKey: generateTestKey(t),
			input:   []byte("test"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CryptoManager{
				Enabled:    tt.enabled,
				privateKey: tt.privKey,
			}

			got, err := cm.DecryptRSA(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptRSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !bytes.Equal(got, tt.input) {
				t.Errorf("DecryptRSA() = %v, want %v", got, tt.input)
			}
		})
	}
}

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Logf("FIPS warning: %v, using nil key", err)
		return nil
	}
	return key
}

func TestDecryptRequest(t *testing.T) {
	tests := []struct {
		name    string
		body    io.Reader
		privKey *rsa.PrivateKey
		wantErr bool
	}{
		{
			name:    "gzip_reader_fail",
			body:    strings.NewReader("not gzip"),
			privKey: nil,
			wantErr: true,
		},
		{
			name:    "gzip_valid_but_short",
			body:    bytes.NewReader([]byte{0x1f, 0x8b, 0x08}),
			privKey: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/", tt.body)

			_, err := decryptRequest(req, tt.privKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v got %v", tt.wantErr, err != nil)
			}
		})
	}
}
