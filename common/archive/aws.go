package archive

import (
	"encoding/json"
	"io/ioutil"
)

type AWSConfig struct {
	Region      string `json:"region"`
	AccessKeyID string `json:"aws_access_key_id"`
	SecretKey   string `json:"aws_secret_access_key"`
	Token       string `json:"aws_token"`
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
