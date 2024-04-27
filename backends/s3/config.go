package s3

import (
	"cas/config"
	"os"
)

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

func (cfg *S3Config) Flags() *config.ConfigGroup {

	group := config.NewConfigGroup("backend: s3")

	group.StringFlag(&cfg.Endpoint, "s3-endpoint", "CAS_S3_ENDPOINT", "", "")
	group.StringFlag(&cfg.Region, "s3-region", "CAS_S3_REGION", "", "")
	group.StringFlag(&cfg.AccessKey, "s3-access-key", "CAS_S3_ACCESS_KEY", "", "")
	group.StringFlag(&cfg.SecretKey, "s3-secret-key", "CAS_S3_SECRET_KEY", "", "")
	group.StringFlag(&cfg.BucketName, "s3-bucket-name", "CAS_S3_BUCKET", "", "")
	group.StringFlag(&cfg.PathPrefix, "s3-path-prefix", "CAS_S3_PATH_PREFIX", "", "")

	return group
}
