package archive

type Archive interface {
	//Remove File: to be implemented
	RemoveFile(filePath string, bucketName string) error

	// UploadFile: upload a local file to a remote destination.
	// The local file name should be passed in as full file Path.
	UploadFile(bucketName string, destinationFolder string, filePath string) error
	// BackupFile: to store a local file onto remote backup location.
	// It also check for file intergrity to ensure upload operation is valid.
	// The local file path should be passed in as full file Path.
	BackupFile(bucketName string, destinationFolder string, filePath string) error
	// CheckFileIntergrity: to ensure that the local file and the upload version is identical.
	CheckFileIntergrity(bucketName string, destinationFolder string, filePath string) (bool, error)
	// GetAuthDataPath: return pre-configured remote folder path to store auth data
	GetAuthDataPath() string
	// GetReserveDataBucketName: return pre-configured remote Bucket to store Reserve Data
	GetReserveDataBucketName() string
	// GetStatDataBucketName: return pre-configured remote Bucket to store stats
	GetStatDataBucketName() string
	// GetPriceAnalyticPath: return pre-configured remote folder Path to store Price Analytic Data
	GetPriceAnalyticPath() string
	// GetLogFolderPath: return pre-configured remote folder to store log
	GetLogFolderPath() string
	// GetLogBucketName: return pre-configure remote bucket to store log
	GetLogBucketName() string
}
