package packager

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ImageMetadata struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Runtime   string    `json:"runtime"`
	BuildTime time.Time `json:"build_time"`
	Size      int64     `json:"size"`
}

type SBOM struct {
	SPDXVersion string      `json:"spdxVersion"`
	Name        string      `json:"name"`
	Created     time.Time   `json:"created"`
	Components  []Component `json:"components"`
}

type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func GenerateKeyPair(configDir string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	keyDir := filepath.Join(configDir, "keys")
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("error creating key directory: %w", err)
	}

	pubKeyPath := filepath.Join(keyDir, "kern.pub")
	privKeyPath := filepath.Join(keyDir, "kern.key")

	if _, err := os.Stat(privKeyPath); err == nil {
		privBytes, err := os.ReadFile(privKeyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading private key: %w", err)
		}
		pubBytes, err := os.ReadFile(pubKeyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading public key: %w", err)
		}
		return ed25519.PublicKey(pubBytes), ed25519.PrivateKey(privBytes), nil
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key pair: %w", err)
	}

	if err := os.WriteFile(privKeyPath, privKey, 0600); err != nil {
		return nil, nil, fmt.Errorf("error saving private key: %w", err)
	}
	if err := os.WriteFile(pubKeyPath, pubKey, 0644); err != nil {
		return nil, nil, fmt.Errorf("error saving public key: %w", err)
	}

	return pubKey, privKey, nil
}

func SignImage(imagePath string, privKey ed25519.PrivateKey) error {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("error reading image: %w", err)
	}

	signature := ed25519.Sign(privKey, data)

	sigPath := imagePath + ".sig"
	if err := os.WriteFile(sigPath, signature, 0644); err != nil {
		return fmt.Errorf("error writing signature: %w", err)
	}

	return nil
}

func VerifySignature(imagePath string, pubKey ed25519.PublicKey) error {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("error reading image: %w", err)
	}

	sigPath := imagePath + ".sig"
	signature, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("error reading signature: %w", err)
	}

	if !ed25519.Verify(pubKey, data, signature) {
		return fmt.Errorf("invalid signature — image may have been tampered with")
	}

	return nil
}

func GenerateSBOM(name, version, runtime string, components []Component) *SBOM {
	return &SBOM{
		SPDXVersion: "SPDX-2.3",
		Name:        name,
		Created:     time.Now(),
		Components:  components,
	}
}

func SaveSBOM(sbom *SBOM, outputPath string) error {
	data, err := json.MarshalIndent(sbom, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing SBOM: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("error saving SBOM: %w", err)
	}

	return nil
}

func GetImageMetadata(name, version, runtime, imagePath string) (*ImageMetadata, error) {
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error reading image: %w", err)
	}

	return &ImageMetadata{
		Name:      name,
		Version:   version,
		Runtime:   runtime,
		BuildTime: time.Now(),
		Size:      info.Size(),
	}, nil
}
