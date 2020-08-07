package gobom

import (
	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	MysqlSource string `json:"mysqlSource"`
	ServerPort string `json:"serverPort"`
}

var appConfig *AppConfig

func InitConfig(path string) error {
	if _, err := toml.DecodeFile(path, &appConfig); err != nil {
		return err
	}
	return nil
}

func GetConfigs() *AppConfig {
	return appConfig
}