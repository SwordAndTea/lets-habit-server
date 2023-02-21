package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
)

// HabitGroup the model to record the related between user and their joined habits
type HabitGroup struct {
	HabitID uint64 `json:"habitID"`
	UID     UID    `json:"uid"`
}

// habitGroupDBHD the handler to operate the habit_group table
type habitGroupDBHD struct{}

// HabitGroupDBHD the default habitGroupDBHD
var HabitGroupDBHD = &habitGroupDBHD{}

// Add insert a HabitGroup record
func (hd *habitGroupDBHD) Add(db *gorm.DB, hg *HabitGroup) response.SError {
	err := db.Create(hg).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add habit group fail")
	}
	return nil
}

// AddMulti insert multiple HabitGroup record at one time
func (hd *habitGroupDBHD) AddMulti(db *gorm.DB, hgs []*HabitGroup) response.SError {
	err := db.Create(hgs).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add multi habit group fail")
	}
	return nil
}

// ListByHabitID list HabitGroup by habit id
func (hd *habitGroupDBHD) ListByHabitID(db *gorm.DB, habitID uint64) ([]*HabitGroup, response.SError) {
	var hgs []*HabitGroup
	err := db.Model(&HabitGroup{}).Where("habit_id=?", habitID).Find(&hgs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit group by habit id fail")
	}
	return hgs, nil
}

// ListByHabitIDs list HabitGroup by a list of habit id
func (hd *habitGroupDBHD) ListByHabitIDs(db *gorm.DB, habitIDs []uint64) ([]*HabitGroup, response.SError) {
	var hgs []*HabitGroup
	err := db.Model(&HabitGroup{}).Where("habit_id in (?)", habitIDs).Find(&hgs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit group by habit ids fail")
	}
	return hgs, nil
}

// ListByUID list HabitGroup by user id
func (hd *habitGroupDBHD) ListByUID(db *gorm.DB, uid UID) ([]*HabitGroup, response.SError) {
	var hgs []*HabitGroup
	err := db.Model(&HabitGroup{}).Where("uid=?", uid).Find(&hgs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit group by uid fail")
	}
	return hgs, nil
}
