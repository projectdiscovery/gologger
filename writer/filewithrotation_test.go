package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectdiscovery/gologger/levels"
)

func newTestFWR(t *testing.T) (*FileWithRotation, string) {
	t.Helper()
	dir := t.TempDir()
	fwr, err := NewFileWithRotation(&FileWithRotationOptions{
		Location: dir,
		FileName: "test.log",
	})
	if err != nil {
		t.Fatalf("NewFileWithRotation: %v", err)
	}
	return fwr, filepath.Join(dir, "test.log")
}

func TestFileWithRotation_WriteDefault(t *testing.T) {
	fwr, path := newTestFWR(t)
	fwr.Write([]byte("hello"), levels.LevelInfo)
	fwr.Close()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.HasPrefix(string(got), "hello") {
		t.Errorf("log content = %q, want prefix %q", got, "hello")
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Errorf("log content not newline terminated: %q", got)
	}
}

func TestFileWithRotation_WriteSilent(t *testing.T) {
	fwr, path := newTestFWR(t)
	fwr.Write([]byte("silent"), levels.LevelSilent)
	fwr.Close()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "silent\n" {
		t.Errorf("log content = %q, want %q", got, "silent\n")
	}
}

func TestFileWithRotation_MissingDirIsCreated(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "logs")
	fwr, err := NewFileWithRotation(&FileWithRotationOptions{
		Location: dir,
		FileName: "test.log",
	})
	if err != nil {
		t.Fatalf("NewFileWithRotation should create nested dirs: %v", err)
	}
	defer fwr.Close()

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir %q to be created: %v", dir, err)
	}
}
