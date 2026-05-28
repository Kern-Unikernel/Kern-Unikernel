package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func secretsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "kern", "secrets")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func deriveKey(name string) []byte {
	hash := sha256.Sum256([]byte("kern-secret-" + name))
	return hash[:]
}

func SetSecret(name, value string) error {
	dir, err := secretsDir()
	if err != nil {
		return fmt.Errorf("error accessing secrets directory: %w", err)
	}

	key := deriveKey(name)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error generating nonce: %w", err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(value), nil)
	secretPath := filepath.Join(dir, name+".enc")

	if err := os.WriteFile(secretPath, []byte(hex.EncodeToString(encrypted)), 0600); err != nil {
		return fmt.Errorf("error saving secret: %w", err)
	}

	fmt.Printf("✓ Secret %q saved\n", name)
	return nil
}

func GetSecret(name string) (string, error) {
	dir, err := secretsDir()
	if err != nil {
		return "", fmt.Errorf("error accessing secrets directory: %w", err)
	}

	secretPath := filepath.Join(dir, name+".enc")
	data, err := os.ReadFile(secretPath)
	if err != nil {
		return "", fmt.Errorf("secret %q not found", name)
	}

	encrypted, err := hex.DecodeString(string(data))
	if err != nil {
		return "", fmt.Errorf("error decoding secret: %w", err)
	}

	key := deriveKey(name)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("invalid secret data")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting secret: %w", err)
	}

	return string(decrypted), nil
}

func ResolveSecrets(envMap map[string]string) (map[string]string, error) {
	resolved := make(map[string]string)
	for k, v := range envMap {
		if strings.HasPrefix(v, "secret:") {
			secretName := strings.TrimPrefix(v, "secret:")
			value, err := GetSecret(secretName)
			if err != nil {
				return nil, fmt.Errorf("could not resolve secret %q: %w", secretName, err)
			}
			resolved[k] = value
		} else {
			resolved[k] = v
		}
	}
	return resolved, nil
}
