package main

import (
	"fmt"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type mockS3Client struct {
	shouldRetry bool
}

const (
	validFileName             = "valid-object.json"
	invalidFileName           = "invalid-object.json"
	nonExistingZip            = "non-existing-zip.zip"
	zipNameForFailingS3Client = "yearly-archives/failing-s3-client.zip"
	invalidZipName            = "failing-s3-client.zip"
)

func (s3Client *mockS3Client) GetObject(bucketName, objectName string) (*minio.Object, error) {
	if s3Client.shouldRetry {
		s3Client.shouldRetry = false
		return nil, fmt.Errorf("network fail")
	}

	if objectName == validFileName {
		return &minio.Object{}, nil
	}

	return nil, fmt.Errorf("Invalid object name")
}
func (s3Client *mockS3Client) PutObject(bucketName, objectName string, reader io.Reader, contentType string) (n int64, err error) {
	fmt.Printf("name is: %s", objectName)
	if objectName == zipNameForFailingS3Client {
		return 0, fmt.Errorf("Cannot upload file.")
	}

	return 0, nil
}
func (s3Client *mockS3Client) ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
	result := make(chan minio.ObjectInfo)
	if objectPrefix == "valid-prefix" {
		go func() {
			result <- minio.ObjectInfo{Key: validFileName}
			close(result)
		}()
	} else {
		close(result)
	}
	return result
}

func TestDownloadFileHappyFlow(t *testing.T) {
	s3Config := newS3Config(&mockS3Client{}, "test-bucket")
	downloadedFile, err := s3Config.downloadFile(validFileName, 2)
	assert.Nil(t, err)
	assert.NotNil(t, downloadedFile)
}

//func TestDownloadFileWithOneRetry(t *testing.T) {
//	initLogs(os.Stdout, os.Stdout, os.Stderr)
//	s3Config := newS3Config(&mockS3Client{shouldRetry:true}, "test-bucket")
//	downloadedFile, err := s3Config.downloadFile(validFileName, 3)
//	assert.Nil(t, err)
//	assert.NotNil(t, downloadedFile)
//}
//
//func TestDownloadFileWithInvalidFileName(t *testing.T) {
//	initLogs(os.Stdout, os.Stdout, os.Stderr)
//	s3Config := newS3Config(&mockS3Client{}, "test-bucket")
//	downloadedFile, err := s3Config.downloadFile(invalidFileName, 3)
//	assert.NotNil(t, err)
//	assert.Nil(t, downloadedFile)
//}

func TestUploadFileHappyFlow(t *testing.T) {
	initLogs(os.Stdout, os.Stdout, os.Stderr)
	zipFile, err := ioutil.TempFile(os.TempDir(), "test.zip")
	assert.Nil(t, err)
	tempZipName := zipFile.Name()
	defer os.Remove(tempZipName)
	zipFile.Close()

	s3Config := newS3Config(&mockS3Client{}, "test-bucket")
	err = s3Config.uploadFile(tempZipName, "test.zip")
	assert.Nil(t, err)
}

func TestUploadFileNonExistingZip(t *testing.T) {
	initLogs(os.Stdout, os.Stdout, os.Stderr)
	s3Config := newS3Config(&mockS3Client{}, "test-bucket")
	err := s3Config.uploadFile(nonExistingZip, "test.zip")
	assert.NotNil(t, err)
}

func TestUploadFileWithS3ClientFailure(t *testing.T) {
	initLogs(os.Stdout, os.Stdout, os.Stderr)
	zipFile, err := ioutil.TempFile(os.TempDir(), "test.zip")
	assert.Nil(t, err)
	tempZipName := zipFile.Name()
	defer os.Remove(tempZipName)
	zipFile.Close()

	s3Config := newS3Config(&mockS3Client{}, "test-bucket")
	err = s3Config.uploadFile(tempZipName, invalidZipName)
	assert.NotNil(t, err)
}