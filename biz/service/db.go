package service

import (
	"context"
	"github.com/swordandtea/fhwh/biz/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var globalDBExecutor *gorm.DB

func InitDB() error {
	db, err := gorm.Open(mysql.Open(config.GlobalConfig.Mysql.DSN))
	if err != nil {
		return err
	}
	globalDBExecutor = db
	return nil
}

func GetDBExecutor() *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return globalDBExecutor.WithContext(ctx)
}
