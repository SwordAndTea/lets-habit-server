package controller

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"gorm.io/gorm"
)

func WithDBTx(dbConn *gorm.DB, dbop func(tx *gorm.DB) response.SError) response.SError {
	if dbConn == nil {
		dbConn = service.GetDBExecutor()
	}
	tx := dbConn.Begin()

	err := dbop(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
