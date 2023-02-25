package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

type HabitCheckRecord struct {
	ID      uint64    `json:"id"`
	HabitID uint64    `json:"habit_id"`
	UID     UID       `json:"uid"`
	CheckAt time.Time `json:"check_at"`
}

type habitCheckRecordDBHD struct{}

var HabitCheckRecordDBHD = &habitCheckRecordDBHD{}

func (hd *habitCheckRecordDBHD) Add(db *gorm.DB, r *HabitCheckRecord) response.SError {
	err := db.Create(r).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one habit check record fail")
	}
	return nil
}

func (hd *habitCheckRecordDBHD) AddMulti(db *gorm.DB, rs []*HabitCheckRecord) response.SError {
	err := db.CreateInBatches(rs, 10).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add multi habit check record fail")
	}
	return nil
}

func (hd *habitCheckRecordDBHD) ListByUID(db *gorm.DB, uid UID, fromTime *time.Time, toTime *time.Time) ([]*HabitCheckRecord, response.SError) {
	q := db.Model(&HabitCheckRecord{}).Where("uid = ?", uid)
	if fromTime != nil {
		q = q.Where("check_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("check_at <= ?", toTime.UTC())
	}
	var results []*HabitCheckRecord
	err := q.Find(&results).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit check records by uid fail")
	}
	return results, nil
}

func (hd *habitCheckRecordDBHD) ListByUIDHabitIDs(db *gorm.DB, uid UID, habitIDs []uint64, fromTime *time.Time, toTime *time.Time) ([]*HabitCheckRecord, response.SError) {
	q := db.Model(&HabitCheckRecord{}).Where("uid = ? AND habit_id in (?)", uid, habitIDs)

	if fromTime != nil {
		q = q.Where("check_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("check_at <= ?", toTime.UTC())
	}
	var results []*HabitCheckRecord
	err := q.Find(&results).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit check records by uid and habit ids fail")
	}
	return results, nil
}
