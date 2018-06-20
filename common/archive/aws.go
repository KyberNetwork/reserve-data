package archive

import (
	"encoding/json"
	"io/ioutil"
)

type AWSConfig struct {
	Region                       string `json:"aws_region"`
	AccessKeyID                  string `json:"aws_access_key_id"`
	SecretKey                    string `json:"aws_secret_access_key"`
	Token                        string `json:"aws_token"`
	ExpiredStatDataBucketName    string `json:"aws_expired_stat_data_bucket_name"`
	ExpiredReserveDataBucketName string `json:"aws_expired_reserve_data_bucket_name"`
	LogBucketName                string `json:"aws_log_bucket_name"`
}

func GetAWSconfigFromFile(path string) (AWSConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return AWSConfig{}, err
	} else {
		result := AWSConfig{}
		err := json.Unmarshal(data, &result)
		return result, err
	}
}
