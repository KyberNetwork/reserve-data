package archive

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Archive struct {
	uploader *s3manager.Uploader
	svc      *s3.S3
}

func (archive *s3Archive) UploadFile(awsfolderPath string, filename string, bucketName string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = archive.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(awsfolderPath + filename),
		Body:   file,
	})

	return err
}

func (archive *s3Archive) CheckFileIntergrity(awsfolderPath string, filename string, bucketName string) (bool, error) {
	//get File info
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return false, err
	}
	fi, err := file.Stat()
	if err != nil {
		return false, err
	}
	//get AWS's file info
	x := s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(awsfolderPath + filename),
	}
	resp, err := archive.svc.ListObjects(&x)
	if err != nil {
		return false, err
	}
	for _, item := range resp.Contents {
		if (*item.Key == filename) && (*item.Size == fi.Size()) {
			return true, nil
		}
	}
	return false, nil
}

func (archive *s3Archive) RemoveFile(filePath string, bucketName string) error {
	var err error
	return err
}

func NewS3Archive(conf AWSConfig) Archive {

	crdtl := credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretKey, conf.Token)
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: crdtl,
	}))
	uploader := s3manager.NewUploader(sess)
	svc := s3.New(sess)
	archive := s3Archive{uploader,
		svc,
	}

	return Archive(&archive)
}
