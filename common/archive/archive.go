package archive

type Archive interface {
	RemoveFile(filePath string, bucketName string) error
	UploadFile(destinationFolder string, fileName string, bucketName string) error
	CheckFileIntergrity(destinationFolder string, fileName string, bucketName string) (bool, error)
}
