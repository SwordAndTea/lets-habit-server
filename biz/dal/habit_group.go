package dal

import (
	"github.com/swordandtea/fhwh/biz/response"
	"gorm.io/gorm"
)

type HabitGroup struct {
	HabitID uint64 `json:"habitID"`
	UID     string `json:"uid"`
}

type habitGroupDBHD struct{}

var HabitGroupDBHD = &habitGroupDBHD{}

func (hd *habitGroupDBHD) Add(db *gorm.DB, hg *HabitGroup) response.SError {
	err := db.Create(hg).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add habit group fail")
	}
	return nil
}

func (hd *habitGroupDBHD) AddMulti(db *gorm.DB, hgs []*HabitGroup) response.SError {
	err := db.Create(hgs).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add multi habit group fail")
	}
	return nil
}

func (hd *habitGroupDBHD) ListByHabitID(db *gorm.DB, habitID uint64) ([]*HabitGroup, response.SError) {
	var hgs []*HabitGroup
	err := db.Where("habit_id=?", habitID).Find(&hgs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit group by habit id fail")
	}
	return hgs, nil
}

func (hd *habitGroupDBHD) ListByUID(db *gorm.DB, uid string) ([]*HabitGroup, response.SError) {
	var hgs []*HabitGroup
	err := db.Where("uid=?", uid).Find(&hgs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit group by uid fail")
	}
	return hgs, nil
}
