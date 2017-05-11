package util

import (
	"fmt"
	"os"
)

// BackupDirectory checks for existance of the $sourceDir and backs it up to backupDir
func BackupDirectory(sourceDir string, backupDir string) (bool, error) {
	_, err := os.Stat(sourceDir)
	var backedup bool
	// Directory exists
	if err == nil {
		if err = os.Rename(sourceDir, backupDir); err != nil {
			return false, fmt.Errorf("Could not back up %q directory: %v", sourceDir, err)
		}
		return true, nil
	} else if !os.IsNotExist(err) { // Directory does not exist but got some other error
		return backedup, fmt.Errorf("Could not determine if %q directory exists: %v", sourceDir, err)
	}
	// Directory does not already exist, nothing to do
	return backedup, nil
}
