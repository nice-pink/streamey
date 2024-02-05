package miniomanager

import (
	"context"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
)

type MinioConfig struct {
	Id       string
	Secret   string
	Endpoint string
}

type MinioManger struct {
	Endpoint string
	client   *minio.Client
}

func NewMinioManager(config MinioConfig, useSSL bool) *MinioManger {
	// Initialize minio client object.
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Id, config.Secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Err(err)
		return nil
	}

	return &MinioManger{client: minioClient, Endpoint: config.Endpoint}
}

func (m *MinioManger) PutFile(bucket string, destFolder string, srcFolder string, srcFilename string, deleteLocal bool) error {
	ctx := context.Background()

	// Upload the test file
	contentType := "application/octet-stream"

	// Upload
	info, err := m.client.FPutObject(ctx, bucket, srcFilename, srcFolder, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("Successfully uploaded ", srcFilename, " of size ", info.Size)

	if deleteLocal {
		filepath := filepath.Join(srcFolder, srcFilename)
		fileErr := filesystem.DeleteFile(filepath)
		if fileErr != nil {
			log.Err(fileErr)
		}
	}

	return nil
}

func (m *MinioManger) DeleteFiles(bucket string, prefix string, olderThanSec int64) bool {
	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range m.client.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		}) {
			if object.Err != nil {
				log.Err(object.Err)
			}
			if filesystem.IsOlderThan(object.LastModified, olderThanSec) {
				log.Info("Delete", object)
				objectsCh <- object
			}
		}
	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	success := true
	for rErr := range m.client.RemoveObjects(context.Background(), bucket, objectsCh, opts) {
		if rErr.Err != nil {
			log.Error(rErr.Err, "Error detected during deletion.")
			success = false
		}
	}
	return success
}
