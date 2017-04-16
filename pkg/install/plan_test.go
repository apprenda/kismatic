package install

import (
	"path/filepath"
	"testing"
)

func TestGenerateAlphaNumericPassword(t *testing.T) {
	_, err := generateAlphaNumericPassword()
	if err != nil {
		t.Error(err)
	}
}

// the test will run either in a container or locally, should always be absolute path
func TestGetSSHKey(t *testing.T) {
	key := getSSHKey()
	if !filepath.IsAbs(key) {
		t.Errorf("Expected key to be an absolute path, instead got %s", key)
	}
}
