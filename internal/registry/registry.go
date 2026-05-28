package registry

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
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

	tag, err := name.NewTag(reference)
	if err != nil {
		return fmt.Errorf("error parsing tag: %w", err)
	}

	img, err := tarball.ImageFromPath(imagePath, &tag)
	if err != nil {
		return fmt.Errorf("error reading image file: %w", err)
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

	if err := tarball.WriteToFile(destPath+"/image.tar", tag, img); err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}

	fmt.Printf("✓ Image pulled from %s\n", reference)
	return nil
}
