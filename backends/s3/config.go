package s3

import "os"

type S3Config struct {
	Endpoint string
	Region   string

	AccessKey string
	SecretKey string

	BucketName string
	PathPrefix string
}

func ConfigFromEnvironment() S3Config {
	return S3Config{
		Endpoint:   os.Getenv("CAS_S3_ENDPOINT"),
		Region:     os.Getenv("CAS_S3_REGION"),
		AccessKey:  os.Getenv("CAS_S3_ACCESS_KEY"),
		SecretKey:  os.Getenv("CAS_S3_SECRET_KEY"),
		BucketName: os.Getenv("CAS_S3_BUCKET"),
		PathPrefix: os.Getenv("CAS_S3_PATH_PREFIX"),
	}
}
