package integration

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKismaticPlatform(t *testing.T) {
	if !testing.Short() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "KismaticPlatform Suite")
	}
}

// Given that kismatic relies on local files, we create a temporary directory
// structure before running tests. The tarball for the kismatic build under test is copied
// to a known location. This location is then used to setup a working directory for each
// test, which will have a pristine copy of kismatic.
// This is what the temp directory structure looks like:
// - $TMP/kismatic-${randomString}
//    - current (contains the tarball for the kismatic build under test)
//    - tests (contains the working directory for each test)
//    - test-resources (contains the test resources that are defined in the suite)
//    - releases (contains subdirectories, one for each downloaded version of kismatic)
//

var kismaticTempDir string
var currentKismaticDir string
var testWorkingDirs string
var testResourcesDir string
var releasesDir string

var _ = BeforeSuite(func() {
	var err error
	kismaticTempDir, err = ioutil.TempDir("", "kismatic-")
	if err != nil {
		Fail(fmt.Sprintf("Failed to make temp dir: %v", err))
	}
	By(fmt.Sprintf("Created temp directory %s", kismaticTempDir))
	// Setup the directory structure
	currentKismaticDir = filepath.Join(kismaticTempDir, "current")
	if err = os.Mkdir(currentKismaticDir, 0700); err != nil {
		Fail(fmt.Sprintf("Failed to make temp dir: %v", err))
	}
	testWorkingDirs = filepath.Join(kismaticTempDir, "tests")
	if err = os.Mkdir(testWorkingDirs, 0700); err != nil {
		Fail(fmt.Sprintf("Failed to make temp dir: %v", err))
	}
	releasesDir = filepath.Join(kismaticTempDir, "releases")
	if err = os.Mkdir(releasesDir, 0700); err != nil {
		Fail(fmt.Sprintf("Failed to make temp dir: %v", err))
	}
	// Copy the current version of kismatic to known location
	cmd := exec.Command("cp", "../out/kismatic.tar.gz", currentKismaticDir)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil {
		Fail("Failed to copy kismatic tarball")
	}
	// Copy test resources to known location. This copy creates the dir.
	testResourcesDir = filepath.Join(kismaticTempDir, "test-resources")
	err = CopyDir("test-resources/", testResourcesDir)
	if err != nil {
		Fail("Failed to copy test resources")
	}
})

var _ = AfterSuite(func() {
	uploadTestLogs()
	if !leaveIt() {
		os.RemoveAll(kismaticTempDir)
	}
})

// sets up a working directory for a test by extracting the kismatic build under
// test to a temp directory. returns the path to the temp directory.
func setupTestWorkingDir() string {
	tmp, err := ioutil.TempDir(testWorkingDirs, "test-")
	if err != nil {
		Fail(fmt.Sprintf("failed to create temp dir: %v", err))
	}
	By("Test working directory is " + tmp)
	err = extractCurrentKismatic(tmp)
	if err != nil {
		Fail(fmt.Sprintf("failed to extract kismatic to %s: %v", tmp, err))
	}
	err = CopyDir(testResourcesDir, filepath.Join(tmp, "test-resources"))
	if err != nil {
		Fail(fmt.Sprintf("failed to copy test resources: %v", err))
	}
	return tmp
}

// sets up a working directory for a test that requires a specific version of kismatic.
// the version of kismatic is extracted into the temp directory.
// returns the path to the temp directory.
func setupTestWorkingDirWithVersion(version string) string {
	tmp, err := ioutil.TempDir(testWorkingDirs, "test-")
	if err != nil {
		Fail(fmt.Sprintf("failed to create temp dir: %v", err))
	}
	By("Test working directory is " + tmp)
	tarball, err := getKismaticReleaseTarball(version)
	if err != nil {
		Fail(fmt.Sprintf("failed to get kismatic tarball for version %s: %v", version, err))
	}
	if err = extractTarball(tarball, tmp); err != nil {
		Fail(fmt.Sprintf("failed to extract kismatic to %s: %v", tmp, err))
	}
	err = CopyDir(testResourcesDir, filepath.Join(tmp, "test-resources"))
	if err != nil {
		Fail(fmt.Sprintf("failed to copy test resources: %v", err))
	}
	return tmp
}

func extractTarball(src, dst string) error {
	return exec.Command("tar", "-zxf", src, "-C", dst).Run()
}

// extracts the current build of kismatic (the one being tested)
func extractCurrentKismatic(dest string) error {
	By(fmt.Sprintf("Extracting current kismatic to directory %q", dest))
	if err := extractTarball(filepath.Join(currentKismaticDir, "kismatic.tar.gz"), dest); err != nil {
		return fmt.Errorf("error extracting kismatic to %s: %v", dest, err)
	}
	return nil
}

// gets the given kismatic release tarball from the local filesystem if available.
// otherwise, it will attempt to download it from github.
func getKismaticReleaseTarball(version string) (string, error) {
	tarFile := filepath.Join(releasesDir, version, "kismatic.tar.gz")
	_, err := os.Stat(tarFile)
	if err == nil {
		// we have already downloaded this release
		return tarFile, nil
	}
	if os.IsNotExist(err) {
		// we haven't downloaded this release. download it.
		if err = os.MkdirAll(filepath.Dir(tarFile), 0700); err != nil {
			return "", fmt.Errorf("failed to create download directory: %v", err)
		}
		if err = downloadKismaticReleaseTarball(version, tarFile); err != nil {
			return "", fmt.Errorf("failed to download ket tarball: %v", err)
		}
		return tarFile, nil
	}
	// some other error ocurred
	return "", fmt.Errorf("failed to stat dir: %v", err)
}

// downloads the specified kismatic version and stores it as file
func downloadKismaticReleaseTarball(version string, file string) error {
	url := fmt.Sprintf("https://github.com/apprenda/kismatic/releases/download/%[1]s/kismatic-%[1]s-linux-amd64.tar.gz", version)
	if runtime.GOOS == "darwin" {
		url = fmt.Sprintf("https://github.com/apprenda/kismatic/releases/download/%[1]s/kismatic-%[1]s-darwin-amd64.tar.gz", version)
	}
	return exec.Command("wget", url, "-O", file).Run()
}

func uploadTestLogs() {
	tests, err := ioutil.ReadDir(testWorkingDirs)
	if err != nil {
		return
	}
	for _, t := range tests {
		if strings.Contains(t.Name(), "test-") {
			uploadKismaticLogs(filepath.Join(testWorkingDirs, t.Name()))
		}
	}
}

// Upload the kismatic package to S3
func uploadKismaticLogs(path string) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKeyID == "" || secretAccessKey == "" {
		return
	}
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion("us-east-1").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	// Tar the folder
	archiveName := "/tmp/kismatic-integration-"
	snapPullRequest := os.Getenv("SNAP_PULL_REQUEST_NUMBER")
	if snapPullRequest != "" {
		archiveName = archiveName + "pr-#" + snapPullRequest + "-"
	}
	snapPipelineCounter := os.Getenv("SNAP_PIPELINE_COUNTER")
	if snapPipelineCounter != "" {
		archiveName = archiveName + "build-#" + snapPipelineCounter + "-"
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	archiveName = archiveName + timestamp + ".tar.gz"

	// upload runs directory
	if _, err := os.Stat(filepath.Join(path, "runs")); os.IsNotExist(err) {
		fmt.Printf("Runs directory not found in %s. Skipping upload of logs to S3.\n", path)
		return
	}
	cmd := exec.Command("tar", "czf", archiveName, filepath.Join(path, "runs"))
	// include the diagnostics directory if exists
	if _, err := os.Stat(filepath.Join(path, "diagnostics")); err == nil {
		fmt.Println("Including diagnostics directory")
		cmd.Args = append(cmd.Args, filepath.Join(path, "diagnostics"))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("failed to tar artifacts")
		return
	}

	file, err := os.Open(archiveName)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	s3path := "/logs/"
	if snapPullRequest != "" {
		s3path = s3path + "pull-requests/" + snapPullRequest + "/"
	}
	if snapPipelineCounter != "" {
		s3path = s3path + snapPipelineCounter + "/"
	}
	s3path = s3path + fileInfo.Name()

	s3bucket := "kismatic-integration-tests"
	downloadPath := fmt.Sprintf("https://console.aws.amazon.com/s3/buckets/%s%s", s3bucket, s3path)
	fmt.Println("Uploading logs to S3. Get them at", downloadPath)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s3bucket),
		Key:           aws.String(s3path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	_, err = svc.PutObject(params)
	if err != nil {
		fmt.Printf("Error uploading logs to S3: %v", err)
	}
}
