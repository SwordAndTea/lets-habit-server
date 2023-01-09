package controller

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"gorm.io/gorm"
)

func WithDBTx(dbop func(tx *gorm.DB) response.SError) response.SError {
	dbHD := service.GetDBExecutor()
	tx := dbHD.Begin()

	err := dbop(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
