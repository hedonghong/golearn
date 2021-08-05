package worker

import (
	"encoding/json"
	"io/ioutil"
)

var (
	GConfig Config
)

type Config struct {
	Endpoints []string `json:"Endpoints"`
	DialTimeout int `json:"DialTimeout"`
	MongodbUri string 	`json:"MongodbUri"`
	MongodbConnectTimeout int `json:"MongodbConnectTimeout"`
	JobLogCommitTimeout int `json:"JobLogCommitTimeout"`
	JobLogBatchSize int `json:"JobLogBatchSize"`
}

func InitConfig(path string) error {
	context, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var config Config

	if err := json.Unmarshal(context, &config); err != nil {
		return err
	}

	GConfig = config
	return nil
}
