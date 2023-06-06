package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

const unconfirmedHabitLogRecordTable = "unconfirmed_habit_log_records"

type unconfirmedHabitLogRecordDBHD struct{}

var UnconfirmedHabitLogRecordDBHD = &unconfirmedHabitLogRecordDBHD{}

func (hd *unconfirmedHabitLogRecordDBHD) Add(db *gorm.DB, r *HabitLogRecord) response.SError {
	err := db.Table(unconfirmedHabitLogRecordTable).Create(r).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one unconfirmed habit log record fail")
	}
	return nil
}

func (hd *unconfirmedHabitLogRecordDBHD) ListByHabitID(db *gorm.DB, habitID uint64, fromTime *time.Time, toTime *time.Time) ([]*HabitLogRecord, response.SError) {
	q := db.Table(unconfirmedHabitLogRecordTable).Where("habit_id = ?", habitID)

	if fromTime != nil {
		q = q.Where("log_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("log_at <= ?", toTime.UTC())
	}
	var results []*HabitLogRecord
	err := q.Find(&results).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list unconfirmed habit log records by uid and habit ids fail")
	}
	return results, nil
}

func (hd *unconfirmedHabitLogRecordDBHD) ListByUIDHabitIDs(db *gorm.DB, uid UID, habitIDs []uint64, fromTime *time.Time, toTime *time.Time) ([]*HabitLogRecord, response.SError) {
	q := db.Table(unconfirmedHabitLogRecordTable).Where("uid = ? AND habit_id in (?)", uid, habitIDs)

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

func (hd *unconfirmedHabitLogRecordDBHD) DeleteByHabitIDAndUID(db *gorm.DB, habitID uint64, uid UID) response.SError {
	err := db.Table(unconfirmedHabitLogRecordTable).Where("habit_id=? and uid=?", habitID, uid).Delete(&HabitLogRecord{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete unconfirmed habit log record fail")
	}
	return nil
}

func (hd *unconfirmedHabitLogRecordDBHD) DeleteByHabitID(db *gorm.DB, habitID uint64, fromTime *time.Time, toTime *time.Time) response.SError {
	q := db.Table(unconfirmedHabitLogRecordTable).Where("habit_id=?", habitID)

	if fromTime != nil {
		q = q.Where("log_at >= ?", fromTime.UTC())
	}
	if toTime != nil {
		q = q.Where("log_at <= ?", toTime.UTC())
	}

	err := q.Delete(&HabitLogRecord{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete unconfirmed habit log record fail")
	}
	return nil
}

func (hd *unconfirmedHabitLogRecordDBHD) DeleteByUID(db *gorm.DB, uid UID) response.SError {
	err := db.Table(unconfirmedHabitLogRecordTable).Where("uid=?", uid).Delete(&HabitLogRecord{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete unconfirmed habit log records by uid fail")
	}
	return nil
}
