package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-token")

	if err := WriteFileAtomic(path, []byte("hello"), 0600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestWriteFileAtomic_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "token")

	if err := WriteFileAtomic(path, []byte("data"), 0600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("expected 'data', got %q", string(data))
	}
}

func TestWriteFileAtomic_Overwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")

	if err := WriteFileAtomic(path, []byte("old"), 0600); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteFileAtomic(path, []byte("new"), 0600); err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("expected 'new', got %q", string(data))
	}
}

func TestWriteKubeconfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")

	if err := WriteKubeconfig(path, "https://k8s.example.com:6443", "/var/run/token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "server: https://k8s.example.com:6443") {
		t.Error("kubeconfig missing server URL")
	}
	if !strings.Contains(content, "tokenFile: /var/run/token") {
		t.Error("kubeconfig missing tokenFile path")
	}
	if !strings.Contains(content, "current-context: default") {
		t.Error("kubeconfig missing current-context")
	}
}
