package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var runtimeOSMap = map[string]string{
	"node":   "unikraft.org/node:21",
	"python": "unikraft.org/python:3.12",
	"go":     "unikraft.org/base:latest",
	"rust":   "unikraft.org/base:latest",
	"bun":    "unikraft.org/bun:1.1",
}

func downloadKernel(runtime, destDir string) (string, error) {
	image, ok := runtimeOSMap[runtime]
	if !ok {
		return "", fmt.Errorf("runtime not supported: %q", runtime)
	}

	// Comprobar caché
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	kernelCacheDir := filepath.Join(cacheDir, "kern", "kernels", runtime)
	cachedKernel := filepath.Join(kernelCacheDir, "kernel")

	if _, err := os.Stat(cachedKernel); err == nil {
		fmt.Printf("→ Using cached kernel for %s\n", runtime)
		destPath := filepath.Join(destDir, "kernel")
		if err := copyFile(cachedKernel, destPath); err != nil {
			return "", err
		}
		return destPath, nil
	}

	// Descargar
	ociDir := filepath.Join(destDir, "oci-"+runtime)
	if err := os.MkdirAll(ociDir, 0755); err != nil {
		return "", err
	}

	fmt.Printf("→ Downloading Unikraft kernel for %s...\n", runtime)

	cmd := exec.Command("skopeo", "copy",
		"--override-arch", "x86_64",
		"--override-os", "qemu",
		"docker://"+image,
		"oci:"+ociDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error downloading kernel: %w", err)
	}

	kernelPath, err := extractKernelFromOCI(ociDir, destDir)
	if err != nil {
		return "", fmt.Errorf("error extracting kernel: %w", err)
	}

	os.RemoveAll(ociDir)

	// Guardar en caché
	if err := os.MkdirAll(kernelCacheDir, 0755); err == nil {
		copyFile(kernelPath, cachedKernel)
		fmt.Printf("✓ Kernel cached at %s\n", kernelCacheDir)
	}

	return kernelPath, nil
}

func extractKernelFromOCI(ociDir, destDir string) (string, error) {
	// Listar blobs
	blobsDir := filepath.Join(ociDir, "blobs", "sha256")
	entries, err := os.ReadDir(blobsDir)
	if err != nil {
		return "", fmt.Errorf("error reading blobs: %w", err)
	}

	kernelPath := filepath.Join(destDir, "kernel")

	// Buscar el blob que contiene el kernel
	for _, entry := range entries {
		blobPath := filepath.Join(blobsDir, entry.Name())

		// Intentar extraer y buscar el kernel
		extractDir := filepath.Join(destDir, "extract-"+entry.Name()[:8])
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			continue
		}

		cmd := exec.Command("tar", "xzf", blobPath, "-C", extractDir, "--no-same-owner", "--no-same-permissions")
		if err := cmd.Run(); err != nil {
			os.RemoveAll(extractDir)
			continue
		}

		// Buscar el kernel dentro
		candidate := filepath.Join(extractDir, "unikraft", "bin", "kernel")
		if _, err := os.Stat(candidate); err == nil {
			// Copiar el kernel al destino
			if err := copyFile(candidate, kernelPath); err != nil {
				os.RemoveAll(extractDir)
				continue
			}
			os.Chmod(kernelPath, 0755)
			os.RemoveAll(extractDir)
			return kernelPath, nil
		}

		os.RemoveAll(extractDir)
	}

	return "", fmt.Errorf("kernel not found in OCI image")
}

func buildRootfs(projectPath, runtime string) (string, error) {
	fmt.Println("→ Building rootfs with buildah...")

	imageName := "kern-rootfs-" + runtime

	// buildah build
	buildCmd := exec.Command("buildah", "build",
		"-t", imageName,
		"-f", filepath.Join(projectPath, "Dockerfile"),
		projectPath,
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		// Fallback a Docker si buildah no está disponible
		fmt.Println("warning: buildah failed, falling back to Docker...")
		return buildRootfsDocker(projectPath, runtime)
	}

	// Exportar imagen a tar OCI
	tarPath := filepath.Join(projectPath, ".unikraft", "rootfs.tar")
	pushCmd := exec.Command("buildah", "push",
		"localhost/"+imageName,
		"oci-archive:"+tarPath,
	)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return "", fmt.Errorf("error exporting rootfs: %w", err)
	}

	// Extraer la capa del OCI tar
	rootfsDir := filepath.Join(projectPath, ".unikraft", "rootfs")
	if err := extractOCILayer(tarPath, rootfsDir); err != nil {
		return "", fmt.Errorf("error extracting rootfs: %w", err)
	}

	// Limpiar
	os.Remove(tarPath)
	exec.Command("buildah", "rmi", "localhost/"+imageName).Run()

	return rootfsDir, nil
}

func buildRootfsDocker(projectPath, runtime string) (string, error) {
	fmt.Println("→ Building rootfs with Docker...")

	imageName := "kern-rootfs-" + runtime

	buildCmd := exec.Command("docker", "build",
		"-t", imageName,
		"-f", filepath.Join(projectPath, "Dockerfile"),
		projectPath,
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("error building rootfs: %w", err)
	}

	rootfsDir := filepath.Join(projectPath, ".unikraft", "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return "", err
	}

	createOut, err := exec.Command("docker", "create",
		"--entrypoint", "", imageName, "sh").Output()
	if err != nil {
		createOut, err = exec.Command("docker", "create",
			imageName, "/bin/true").Output()
		if err != nil {
			return "", fmt.Errorf("error creating container: %w", err)
		}
	}

	containerID := strings.TrimSpace(string(createOut))

	exportCmd := exec.Command("docker", "export", containerID)
	tarCmd := exec.Command("tar", "-x", "-C", rootfsDir)
	tarCmd.Stdin, _ = exportCmd.StdoutPipe()
	tarCmd.Stderr = os.Stderr

	if err := tarCmd.Start(); err != nil {
		return "", err
	}
	if err := exportCmd.Run(); err != nil {
		return "", err
	}
	if err := tarCmd.Wait(); err != nil {
		return "", err
	}

	exec.Command("docker", "rm", containerID).Run()
	return rootfsDir, nil
}

func extractOCILayer(tarPath, rootfsDir string) error {
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return err
	}

	// Leer el index.json del tar OCI
	indexCmd := exec.Command("tar", "xOf", tarPath, "index.json")
	indexData, err := indexCmd.Output()
	if err != nil {
		return fmt.Errorf("error reading OCI index: %w", err)
	}

	var index struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("error parsing OCI index: %w", err)
	}

	if len(index.Manifests) == 0 {
		return fmt.Errorf("no manifests found in OCI image")
	}

	manifestDigest := strings.TrimPrefix(index.Manifests[0].Digest, "sha256:")
	manifestPath := "blobs/sha256/" + manifestDigest

	manifestCmd := exec.Command("tar", "xOf", tarPath, manifestPath)
	manifestData, err := manifestCmd.Output()
	if err != nil {
		return fmt.Errorf("error reading manifest: %w", err)
	}

	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("error parsing manifest: %w", err)
	}

	if len(manifest.Layers) == 0 {
		return fmt.Errorf("no layers found in OCI image")
	}

	layerDigest := strings.TrimPrefix(manifest.Layers[0].Digest, "sha256:")
	layerPath := "blobs/sha256/" + layerDigest

	// Extraer la capa
	extractCmd := exec.Command("tar", "xOf", tarPath, layerPath)
	untarCmd := exec.Command("tar", "-xz", "-C", rootfsDir, "--no-same-owner", "--no-same-permissions")
	layer, err := extractCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error opening OCI layer: %w", err)
	}
	untarCmd.Stdin = layer
	untarCmd.Stderr = os.Stderr
	if err := untarCmd.Start(); err != nil {
		return fmt.Errorf("error starting layer extraction: %w", err)
	}
	if err := extractCmd.Run(); err != nil {
		_ = untarCmd.Process.Kill()
		_ = untarCmd.Wait()
		return fmt.Errorf("error reading OCI layer: %w", err)
	}
	if err := untarCmd.Wait(); err != nil {
		return fmt.Errorf("error extracting layer: %w", err)
	}

	return nil
}

func generateInitramfs(rootfsDir, outputPath string) error {
	fmt.Println("→ Generating initramfs CPIO...")

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating initramfs: %w", err)
	}
	defer output.Close()

	findCmd := exec.Command("find", ".")
	findCmd.Dir = rootfsDir
	cpioCmd := exec.Command("cpio", "-o", "-H", "newc")
	files, err := findCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error listing rootfs: %w", err)
	}
	cpioCmd.Stdin = files
	cpioCmd.Stdout = output
	cpioCmd.Stderr = os.Stderr
	if err := cpioCmd.Start(); err != nil {
		return fmt.Errorf("error starting cpio: %w", err)
	}
	if err := findCmd.Run(); err != nil {
		_ = cpioCmd.Process.Kill()
		_ = cpioCmd.Wait()
		return fmt.Errorf("error listing rootfs: %w", err)
	}
	if err := cpioCmd.Wait(); err != nil {
		return fmt.Errorf("error generating initramfs: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, 0755)
}
