package controller

import (
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
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
