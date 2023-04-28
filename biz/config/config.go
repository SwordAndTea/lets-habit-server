package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	RunModeLocal = "local"
	RunModelTest = "test"
	RunModelProd = "prod"
)

type LogConfig struct {
	Level string `yaml:"level" json:"level"`
}

type MysqlConfig struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

type ObjectStorageConfig struct {
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Region   string `yaml:"region" json:"region"`
	AK       string `yaml:"ak" json:"ak"`
	SK       string `yaml:"sk" json:"sk"`
	Bucket   string `yaml:"bucket" json:"bucket"`
}

type EmailServiceConfig struct {
	Sender        string `yaml:"sender" json:"sender"`
	Host          string `yaml:"host" json:"host"`
	Port          uint32 `yaml:"port" json:"port"`
	AuthCode      string `yaml:"auth_code" json:"auth_code"`
	ActivateURI   string `yaml:"activate_uri" json:"activate_uri"`
	ActivateParam string `yaml:"activate_param" json:"activate_param"`
	BindURI       string `yaml:"bind_uri" json:"bind_uri"`
	BindParam     string `yaml:"bind_param" json:"bind_param"`
}

type JWTConfig struct {
	Cypher string `yaml:"cypher" json:"cypher"`
}

type RuntimeConfig struct {
	RunMode       string              `yaml:"-" json:"-"`
	Log           LogConfig           `yaml:"log" json:"log"`
	Mysql         MysqlConfig         `yaml:"mysql" json:"mysql"`
	ObjectStorage ObjectStorageConfig `yaml:"object_storage" json:"object_storage"`
	EmailService  EmailServiceConfig  `yaml:"email_service" json:"email_service"`
	JWT           JWTConfig           `yaml:"jwt" json:"jwt"`
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
	c.RunMode = runMode
	GlobalConfig = c

	// read some env to overwrite config, usually is some sensitive config
	if mysqlDSN := os.Getenv("MYSQL_DSN"); mysqlDSN != "" {
		GlobalConfig.Mysql.DSN = mysqlDSN
	}

	if emailServiceSender := os.Getenv("EMAIL_SERVICE_SENDER"); emailServiceSender != "" {
		GlobalConfig.EmailService.Sender = emailServiceSender
	}

	if emailServiceAuthCode := os.Getenv("EMAIL_SERVICE_AUTH_CODE"); emailServiceAuthCode != "" {
		GlobalConfig.EmailService.AuthCode = emailServiceAuthCode
	}

	if jwtCypher := os.Getenv("JWT_CYPHER"); jwtCypher != "" {
		GlobalConfig.JWT.Cypher = jwtCypher
	}

	return nil
}
