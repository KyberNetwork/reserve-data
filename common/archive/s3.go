package archive

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	REGION = "ap-southeast-1"
)

type s3Archive struct {
	uploader *s3manager.Uploader
	svc      *s3.S3
}

func (archive *s3Archive) UploadFile(filepath string, filename string, bucketName string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = archive.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath),
		Body:   file,
	})

	x := s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(filepath),
	}
	resp, err := archive.svc.ListObjects(&x)
	if err != nil {
		return err
	}
	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}

	fi, err := file.Stat()
	if err != nil {
		log.Printf("cunt")
	}
	log.Printf("\n\n\n\n\n%d\n\n\n\n", fi.Size())
	return err
}

func (archive *s3Archive) RemoveFile(filePath string, bucketName string) error {
	var err error
	return err
}

func NewS3Archive() Archive {

	crdtl := credentials.NewStaticCredentials("AKIAIOEEV6TUEYCP4N6Q",
		"eKusKOt4aM/0fpAoaSzX/SyW0wpN6Hjc9biWPHjO",
		"")
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(REGION),
		Credentials: crdtl,
	}))
	uploader := s3manager.NewUploader(sess)
	svc := s3.New(sess)
	archive := s3Archive{uploader,
		svc,
	}

	return Archive(&archive)
}
