package archive

import (
	"fmt"
	"os"
	"strings"

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
	awsConf  AWSConfig
}

func (archive *s3Archive) BackupFile(bucketName string, destinationFolder string, filePath string) error {
	err := archive.UploadFile(bucketName, destinationFolder, filePath)
	if err != nil {
		return err
	}
	intergrity, err := archive.CheckFileIntergrity(bucketName, destinationFolder, filePath)
	if err != nil {
		return err
	}
	if !intergrity {
		return fmt.Errorf("Archive: Upload File  %s: corrupted", filePath)
	}
	return nil
}

func enforceFolderPath(fp string) string {
	if string(fp[len(fp)-1]) != "/" {
		fp = fp + "/"
	}
	return fp
}
func (archive *s3Archive) UploadFile(bucketName string, awsfolderPath string, filePath string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = archive.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(enforceFolderPath(awsfolderPath) + getFileNameFromFilePath(filePath)),
		Body:   file,
	})

	return err
}

func getFileNameFromFilePath(filePath string) string {
	elems := strings.Split(filePath, "/")
	fileName := elems[len(elems)-1]
	return fileName
}

func (archive *s3Archive) CheckFileIntergrity(bucketName string, awsfolderPath string, filePath string) (bool, error) {
	//get File info
	file, err := os.Open(filePath)
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
		Prefix: aws.String(enforceFolderPath(awsfolderPath) + getFileNameFromFilePath(filePath)),
	}
	resp, err := archive.svc.ListObjects(&x)
	if err != nil {
		return false, err
	}

	for _, item := range resp.Contents {
		remoteFileName := getFileNameFromFilePath(*item.Key)
		localFileName := getFileNameFromFilePath(filePath)
		if (remoteFileName == localFileName) && (*item.Size == fi.Size()) {
			return true, nil
		}
	}
	return false, nil
}

func (archive *s3Archive) RemoveFile(filePath string, bucketName string) error {
	var err error
	return err
}

func (archive *s3Archive) GetAuthDataPath() string {
	return archive.awsConf.ExpiredAuthDataFolderPath

}

func (archive *s3Archive) GetReserveDataBucketName() string {
	return archive.awsConf.ExpiredReserveDataBucketName
}

//GetStatDataBucketName returns the bucket in which the backup Data is stored.
//This should be passed in from JSON configure file
func (archive *s3Archive) GetStatDataBucketName() string {
	return archive.awsConf.ExpiredStatDataBucketName
}

//GetPriceAnalyticPath returns the folder path to store Expired Price Analytic Data.
//Ths should be passed in from JSON configure file
func (archive *s3Archive) GetPriceAnalyticPath() string {
	return archive.awsConf.ExpiredPriceAnalyticFolderPath
}

func (archive *s3Archive) GetLogFolderPath() string {
	return archive.awsConf.LogFolderPath
}

func (archive *s3Archive) GetLogBucketName() string {
	return archive.awsConf.LogBucketName
}

func NewS3Archive(conf AWSConfig) *s3Archive {

	crdtl := credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretKey, conf.Token)
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: crdtl,
	}))
	uploader := s3manager.NewUploader(sess)
	svc := s3.New(sess)
	archive := s3Archive{uploader,
		svc,
		conf,
	}

	return &archive
}
