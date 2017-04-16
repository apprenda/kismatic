package util

import "os"

// RunningInDocker stat /.dockerenv file that should be available in all docker containers
func RunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
