package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const previousKismaticVersion = "v1.0.0"

func TestKismaticPlatform(t *testing.T) {
	if !testing.Short() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "KismaticPlatform Suite")
	}
}

var kisPath string
var kisReleasedPath string
var _ = BeforeSuite(func() {
	var err error
	kisPath, err = ExtractKismaticToTemp()
	if err != nil {
		Fail("Failed to extract kismatic")
	}
	err = CopyDir("test-tls/", filepath.Join(kisPath, "test-tls"))
	if err != nil {
		Fail("Failed to copy test certs")
	}
	// setup previous version of Kismatic
	kisReleasedPath, err = DownloadKismaticRelease(previousKismaticVersion)
	if err != nil {
		Fail("Failed to download kismatic released")
	}
})

var _ = AfterSuite(func() {
	if !leaveIt() {
		os.RemoveAll(kisPath)
		os.RemoveAll(kisReleasedPath)
	}
})
