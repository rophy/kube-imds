package client

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically by writing to a temp file
// in the same directory and then renaming.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp.*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	defer func() {
		if tmpName != "" {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}

	tmpName = "" // prevent deferred cleanup
	return nil
}

// WriteKubeconfig writes a kubeconfig YAML file that uses tokenFile-based auth.
func WriteKubeconfig(path, kubeAPIServer, tokenPath string) error {
	kubeconfig := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
  - cluster:
      server: %s
    name: default
contexts:
  - context:
      cluster: default
      user: default
    name: default
current-context: default
users:
  - name: default
    user:
      tokenFile: %s
`, kubeAPIServer, tokenPath)

	return WriteFileAtomic(path, []byte(kubeconfig), 0600)
}
