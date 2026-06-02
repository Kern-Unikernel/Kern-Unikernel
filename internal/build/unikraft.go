package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	ociDir := filepath.Join(destDir, "oci-"+runtime)
	if err := os.MkdirAll(ociDir, 0755); err != nil {
		return "", err
	}

	fmt.Printf("→ Downloading Unikraft kernel for %s...\n", runtime)

	// Descargar imagen OCI con skopeo
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

	// Extraer el kernel del blob OCI
	kernelPath, err := extractKernelFromOCI(ociDir, destDir)
	if err != nil {
		return "", fmt.Errorf("error extracting kernel: %w", err)
	}

	// Limpiar directorio OCI
	os.RemoveAll(ociDir)

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

		cmd := exec.Command("tar", "xzf", blobPath, "-C", extractDir)
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
	fmt.Println("→ Building rootfs with Docker...")

	imageName := "kern-rootfs-" + runtime

	// Docker build
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

	// Extraer el contenido del contenedor
	rootfsDir := filepath.Join(projectPath, ".unikraft", "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return "", err
	}

	// Crear contenedor temporal
	createOut, err := exec.Command("docker", "create",
		"--entrypoint", "",
		imageName, "sh").Output()
	if err != nil {
		// Intentar sin entrypoint
		createOut, err = exec.Command("docker", "create",
			imageName, "/bin/true").Output()
		if err != nil {
			return "", fmt.Errorf("error creating container: %w", err)
		}
	}

	containerID := string(createOut)
	if len(containerID) > 0 && containerID[len(containerID)-1] == '\n' {
		containerID = containerID[:len(containerID)-1]
	}

	// Exportar y extraer
	exportCmd := exec.Command("docker", "export", containerID)
	tarCmd := exec.Command("tar", "-x", "-C", rootfsDir)
	tarCmd.Stdin, _ = exportCmd.StdoutPipe()
	tarCmd.Stdout = os.Stdout
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

	// Eliminar contenedor
	exec.Command("docker", "rm", containerID).Run()

	return rootfsDir, nil
}

func generateInitramfs(rootfsDir, outputPath string) error {
	fmt.Println("→ Generating initramfs CPIO...")

	cpioCmd := exec.Command("sh", "-c",
		fmt.Sprintf("cd %s && find . | cpio -o -H newc > %s", rootfsDir, outputPath))
	cpioCmd.Stdout = os.Stdout
	cpioCmd.Stderr = os.Stderr

	if err := cpioCmd.Run(); err != nil {
		return fmt.Errorf("error generating initramfs: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
