package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

type HabitLogRecord struct {
	ID      uint64    `json:"id"`
	HabitID uint64    `json:"habit_id"`
	UID     UID       `json:"uid"`
	LogAt   time.Time `json:"log_at"`
}

type habitLogRecordDBHD struct{}

var HabitLogRecordDBHD = &habitLogRecordDBHD{}

func (hd *habitLogRecordDBHD) Add(db *gorm.DB, r *HabitLogRecord) response.SError {
	err := db.Create(r).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one habit log record fail")
	}
	return nil
}

func (hd *habitLogRecordDBHD) AddMulti(db *gorm.DB, rs []*HabitLogRecord) response.SError {
	err := db.CreateInBatches(rs, 10).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add multi habit log record fail")
	}
	return nil
}

func (hd *habitLogRecordDBHD) ListByUID(db *gorm.DB, uid UID, fromTime *time.Time, toTime *time.Time) ([]*HabitLogRecord, response.SError) {
	q := db.Model(&HabitLogRecord{}).Where("uid = ?", uid)
	if fromTime != nil {
		q = q.Where("log_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("log_at <= ?", toTime.UTC())
	}
	var results []*HabitLogRecord
	err := q.Find(&results).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit log records by uid fail")
	}
	return results, nil
}

func (hd *habitLogRecordDBHD) ListByUIDHabitIDs(db *gorm.DB, uid UID, habitIDs []uint64, fromTime *time.Time, toTime *time.Time) ([]*HabitLogRecord, response.SError) {
	q := db.Model(&HabitLogRecord{}).Where("uid = ? AND habit_id in (?)", uid, habitIDs)

	if fromTime != nil {
		q = q.Where("log_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("log_at <= ?", toTime.UTC())
	}
	var results []*HabitLogRecord
	err := q.Find(&results).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list habit log records by uid and habit ids fail")
	}
	return results, nil
}

func (hd *habitLogRecordDBHD) DeleteByHabitIDAndUID(db *gorm.DB, habitID uint64, uid UID) response.SError {
	err := db.Where("habit_id=? and uid=?", habitID, uid).Delete(&HabitLogRecord{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete habit log record fail")
	}
	return nil
}

func (hd *habitLogRecordDBHD) DeleteByUID(db *gorm.DB, uid UID) response.SError {
	err := db.Where("uid=?", uid).Delete(&HabitLogRecord{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete habit log records by uid fail")
	}
	return nil
}
