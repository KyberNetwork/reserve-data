package configuration

import (
	"encoding/json"
	"io/ioutil"
)

type BlockSetRate struct {
	BeginBlockSetRate uint64 `json:"begin_block_setrate"`
}

func GetBeginBlockSetRate(path string) uint64 {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result := BlockSetRate{}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		panic(err)
	}
	beginBlockSetRate := result.BeginBlockSetRate
	return beginBlockSetRate
}
