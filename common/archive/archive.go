package archive

type Archive interface {
	RemoveFile(bucketName string, destinationFolder string, filePath string) error

	// UploadFile: upload a local file to a remote destination.
	// The local file name should be passed in as full file Path.
	UploadFile(bucketName string, destinationFolder string, filePath string) error

	// CheckFileIntergrity: to ensure that the local file and the upload version is identical.
	CheckFileIntergrity(bucketName string, destinationFolder string, filePath string) (bool, error)

	// GetReserveDataBucketName: return pre-configured remote Bucket to store Reserve Data
	GetReserveDataBucketName() string

	// GetStatDataBucketName: return pre-configured remote Bucket to store stats
	GetStatDataBucketName() string

	// GetLogBucketName: return pre-configured remote Bucket to store logs
	GetLogBucketName() string
}
