package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

const HabitLogDelayHours = 4

type CheckDay uint8

const (
	CheckDaySunday CheckDay = 1 << iota
	CheckDayMonday
	CheckDayTuesday
	CheckDayWednesday
	CheckDayThursday
	CheckDayFriday
	CheckDaySaturday
	CheckDayAll = CheckDayMonday | CheckDayTuesday | CheckDayWednesday | CheckDayThursday |
		CheckDayFriday | CheckDaySaturday | CheckDaySunday
)

func (c CheckDay) IsValid() bool {
	return c > 0 && c <= CheckDayAll
}

func (c CheckDay) Has(d CheckDay) bool {
	return c&d > 0
}

// Habit the habit model to represent a habit
type Habit struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	IdentityToForm *string   `json:"identity_to_form"`
	LogDays        CheckDay  `json:"log_days"`
	Owner          UID       `json:"owner"`
	CreateAt       time.Time `json:"create_at"`
}

// habitDBHD the handler to operate the habit table
type habitDBHD struct{}

// HabitDBHD the default habitDBHD
var HabitDBHD = &habitDBHD{}

// Add insert a Habit record into db
func (hd *habitDBHD) Add(db *gorm.DB, h *Habit) response.SError {
	err := db.Create(h).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one habit fail")
	}
	return nil
}

// GetByID get ad Habit by id
func (hd *habitDBHD) GetByID(db *gorm.DB, id uint64) (*Habit, response.SError) {
	var h *Habit
	err := db.Where("id=?", id).First(&h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get habit by id fail")
	}
	return h, nil
}

// ListUserJoinedHabits list all Habits one user joined
func (hd *habitDBHD) ListUserJoinedHabits(db *gorm.DB, uid UID, pagination *Pagination) ([]*Habit, uint, response.SError) {
	var hs []*Habit

	subquery := db.Model(&HabitGroup{}).Select("habit_id").Where("uid=?", uid)
	var count int64
	err := db.Model(&Habit{}).Where("id in (?)", subquery).Count(&count).Error
	if err != nil {
		return nil, 0, response.ErrroCode_InternalUnknownError.Wrap(err, "list user joined habits fail")
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	err = db.Where("id in (?)", subquery).Offset(int(offset)).Limit(int(pagination.PageSize)).Find(&hs).Error
	if err != nil {
		return nil, 0, response.ErrroCode_InternalUnknownError.Wrap(err, "list user joined habits fail")
	}
	return hs, uint(count), nil
}

type HabitUpdatableFields struct {
	Owner *UID
}

func (hd *habitDBHD) UpdateHabit(db *gorm.DB, id uint64, updateFields *HabitUpdatableFields) response.SError {
	updates := map[string]interface{}{}
	if updateFields.Owner != nil {
		updates["owner"] = *updateFields.Owner
	}
	if len(updates) == 0 {
		return nil
	}
	err := db.Model(&Habit{}).Where("uid=?", id).Updates(updates).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update habit fail")
	}
	return nil
}

// DeleteByID delete a Habit record from db by id
func (hd *habitDBHD) DeleteByID(db *gorm.DB, id uint64) response.SError {
	err := db.Where("id=?", id).Delete(&Habit{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete habit fail")
	}
	return nil
}
