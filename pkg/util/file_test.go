package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupDirectoryExists(t *testing.T) {
	sourceDir := filepath.Join("/tmp", ".helm")
	err := os.Mkdir(sourceDir, 0755)
	if err != nil {
		t.Errorf("Expected error creating /tmp to be nil, got: %v", err)
	}

	exists, err := BackupDirectory(sourceDir, filepath.Join("/tmp", ".helm.bak"))
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	}
	if !exists {
		t.Errorf("Expected directory to exist")
	}
}

func TestBackupClientDirectoryNotExists(t *testing.T) {
	exists, err := BackupDirectory(filepath.Join("/tmp", ".helm"), filepath.Join("/tmp", ".helm.bak"))
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	}
	if exists {
		t.Errorf("Expected directory to not exist")
	}
}
