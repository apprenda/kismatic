package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKismaticPlatform(t *testing.T) {
	if !testing.Short() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "KismaticPlatform Suite")
	}
}

var kisPath string
var _ = BeforeSuite(func() {
	var err error
	kisPath, err = ExtractKismaticToTemp()
	if err != nil {
		Fail("Failed to extract kismatic")
	}
	err = CopyDir("test-resources/", filepath.Join(kisPath, "test-resources"))
	if err != nil {
		Fail("Failed to copy test certs")
	}
})

var _ = AfterSuite(func() {
	uploadKismaticLogs(kisPath)
	if !leaveIt() {
		os.RemoveAll(kisPath)
	}
})

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
		fmt.Println("Runs directory not found. Skipping upload of logs to S3.")
		return
	}
	cmd := exec.Command("tar", "czf", archiveName, filepath.Join(path, "runs"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		fmt.Printf("failed to tar artifacts")
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

	fmt.Println("Uploading logs to S3 at", s3path)
	params := &s3.PutObjectInput{
		Bucket:        aws.String("kismatic-integration-tests"),
		Key:           aws.String(s3path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	resp, err := svc.PutObject(params)
	if err != nil {
		fmt.Printf("bad response: %s", err)
	}
	fmt.Printf("response %s", awsutil.StringValue(resp))
}
