package archive

type Archive interface {
	RemoveFile(filePath string, bucketName string) error
	UploadFile(filePath string, fileName string, bucketName string) error
}
