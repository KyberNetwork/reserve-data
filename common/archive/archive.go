package archive

type Archive interface {
	RemoveFile(bucketName string, destinationFolder string, filePath string) error
	UploadFile(bucketName string, destinationFolder string, filePath string) error
	CheckFileIntergrity(bucketName string, destinationFolder string, filePath string) (bool, error)
	GetReserveDataBucketName() string
	GetStatDataBucketName() string
}
