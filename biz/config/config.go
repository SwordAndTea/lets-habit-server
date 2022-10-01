package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type LogConfig struct {
	Level string `yaml:"level" json:"level"`
}

type MysqlConfig struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

type RuntimeConfig struct {
	Log   LogConfig   `yaml:"log" json:"log"`
	Mysql MysqlConfig `yaml:"mysql" json:"mysql"`
}

var GlobalConfig *RuntimeConfig

func InitConfig(filePath string) error {
	configData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var allConf map[string]*RuntimeConfig
	err = yaml.Unmarshal(configData, &allConf)
	if err != nil {
		return err
	}

	runMode := os.Getenv("RUN_MODE")
	c, ok := allConf[runMode]
	if !ok {
		return fmt.Errorf("unknown RUN_MODE %s", runMode)
	}
	GlobalConfig = c

	// read some env to overwrite config
	if mysqlDSN := os.Getenv("MYSQL_DSN"); mysqlDSN != "" {
		GlobalConfig.Mysql.DSN = mysqlDSN
	}

	return nil
}
