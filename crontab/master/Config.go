package master

import (
	"encoding/json"
	"io/ioutil"
)

var (
	GConfig Config
)

type Config struct {
	ApiPort int `json:"ApiPort"`
	ApiReadTimeOut int `json:"ApiReadTimeOut"`
	ApiWriteTimeOut int `json:"ApiWriteTimeOut"`
	Endpoints []string `json:"Endpoints"`
	DialTimeout int `json:"DialTimeout"`
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
