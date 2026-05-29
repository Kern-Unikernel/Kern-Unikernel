package registry

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"path/filepath"
)

type Credentials struct {
	Username string
	Password string
}

func PushImage(imagePath, reference string, creds Credentials) error {
	ref, err := name.ParseReference(reference)
	if err != nil {
		return fmt.Errorf("invalid image reference %q: %w", reference, err)
	}

	// Crear una capa OCI desde el archivo .uki
	layer, err := tarball.LayerFromFile(imagePath)
	if err != nil {
		return fmt.Errorf("error creating layer from file: %w", err)
	}

	// Añadir la capa a una imagen OCI vacía
	img, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("error creating image: %w", err)
	}

	auth := authn.FromConfig(authn.AuthConfig{
		Username: creds.Username,
		Password: creds.Password,
	})

	if err := remote.Write(ref, img, remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	fmt.Printf("✓ Image pushed to %s\n", reference)
	return nil
}

func PullImage(reference, destPath string, creds Credentials) error {
	ref, err := name.ParseReference(reference)
	if err != nil {
		return fmt.Errorf("invalid image reference %q: %w", reference, err)
	}

	auth := authn.FromConfig(authn.AuthConfig{
		Username: creds.Username,
		Password: creds.Password,
	})

	img, err := remote.Image(ref, remote.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	tag, err := name.NewTag(reference)
	if err != nil {
		return fmt.Errorf("error parsing tag: %w", err)
	}

	ukiPath := filepath.Join(destPath, "image.uki")
	if err := tarball.WriteToFile(ukiPath, tag, img); err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}

	fmt.Printf("✓ Image pulled to %s\n", ukiPath)
	return nil
}
