package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupClientDirectoryExists(t *testing.T) {
	err := os.Mkdir(filepath.Join("/tmp", ".helm"), 0755)
	if err != nil {
		t.Errorf("Expected error creating /tmp to be nil, got: %v", err)
	}
	helm, err := DefaultHelmClient()
	if err != nil {
		t.Errorf("Expected error geting client to be nil, got: %v", err)
	}
	helm.ClientDirectory = filepath.Join("/tmp", ".helm")

	exists, err := helm.BackupClient()
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	}
	if !exists {
		t.Errorf("Expected directory to exist")
	}
}

func TestBackupClientDirectoryNotExists(t *testing.T) {
	helm, err := DefaultHelmClient()
	if err != nil {
		t.Errorf("Expected error geting client to be nil, got: %v", err)
	}
	helm.ClientDirectory = filepath.Join("/tmp", ".dne")

	exists, err := helm.BackupClient()
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	}
	if exists {
		t.Errorf("Expected directory to not exist")
	}
}
