package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/fhwh/biz/response"
	"gorm.io/gorm"
	"time"
)

type HabitCheckType string

const (
	HabitCheckTypeBinary   HabitCheckType = "binary"
	HabitCheckTimeInterval HabitCheckType = "time_interval"
)

func (t HabitCheckType) IsValid() bool {
	switch t {
	case HabitCheckTypeBinary, HabitCheckTimeInterval:
		return true
	default:
		return false
	}
}

type HabitCheckFrequency string

const (
	HabitCheckFrequencyDaily   HabitCheckFrequency = "daily"
	HabitCheckFrequencyWeekly  HabitCheckFrequency = "weekly"
	HabitCheckFrequencyMonthly HabitCheckFrequency = "monthly"
)

func (f HabitCheckFrequency) IsValid() bool {
	switch f {
	case HabitCheckFrequencyDaily, HabitCheckFrequencyWeekly, HabitCheckFrequencyMonthly:
		return true
	default:
		return false
	}
}

// Habit the habit model to represent a habit
type Habit struct {
	ID                 uint64              `json:"id"`
	Creator            string              `json:"creator"`
	CreateAt           time.Time           `json:"create_at"`
	Name               string              `json:"content"`
	CheckType          HabitCheckType      `json:"check_type"`
	CheckFrequency     HabitCheckFrequency `json:"check_frequency"`
	CheckDeadlineDelay time.Duration       `json:"check_deadline_delay"`
}

// habitDBHD habit db handler
type habitDBHD struct{}

var HabitDBHD = &habitDBHD{}

func (hd *habitDBHD) Add(db *gorm.DB, h *Habit) response.SError {
	err := db.Create(h).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one habit fail")
	}
	return nil
}

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

func (hd *habitDBHD) DeleteByID(db *gorm.DB, id uint64) response.SError {
	err := db.Where("id=?", id).Delete(&Habit{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete habit fail")
	}
	return nil
}
