package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

// HabitCheckType the type indicate what the habit's check type is
// HabitCheckTypeBinary means the habit is only got to be checked
// HabitCheckTimeInterval means the user can fill the checked the habit
// and fill how much time they used in this habit
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

// HabitCheckFrequency the frequency the user need to check the habit
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

// HabitPublicLevel the public level of a habit
// HabitPublicLevelPublic means every one can see this habit
// HabitPublicLevelPrivate means only users that joined this habit can see this habit
type HabitPublicLevel string

const (
	HabitPublicLevelPublic  HabitPublicLevel = "public"
	HabitPublicLevelPrivate HabitPublicLevel = "private"
)

func (l HabitPublicLevel) IsValid() bool {
	switch l {
	case HabitPublicLevelPublic, HabitPublicLevelPrivate:
		return true
	default:
		return false
	}
}

type HabitCheckDeadlineDelay time.Duration

const (
	HabitCheckDeadlineDelayOneHour  = HabitCheckDeadlineDelay(time.Hour)
	HabitCheckDeadlineDelayTwoHour  = HabitCheckDeadlineDelay(time.Hour * 2)
	HabitCheckDeadlineDelayFourHour = HabitCheckDeadlineDelay(time.Hour * 4)
)

func (d HabitCheckDeadlineDelay) IsValid() bool {
	switch d {
	case HabitCheckDeadlineDelayOneHour, HabitCheckDeadlineDelayTwoHour, HabitCheckDeadlineDelayFourHour:
		return true
	default:
		return false
	}
}

// Habit the habit model to represent a habit
type Habit struct {
	ID                 uint64                  `json:"id"`
	Creator            UID                     `json:"creator"`
	CreateAt           time.Time               `json:"create_at"`
	Name               string                  `json:"content"`
	PublicLevel        HabitPublicLevel        `json:"public_level"`
	CheckType          HabitCheckType          `json:"check_type"`
	CheckFrequency     HabitCheckFrequency     `json:"check_frequency"`
	CheckDeadlineDelay HabitCheckDeadlineDelay `json:"check_deadline_delay"`
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

	subquery := db.Table("habit_group").Select("habit_id").Where("uid=?", uid)
	var count int64
	err := db.Where("id in (?)", subquery).Count(&count).Error
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

// DeleteByID delete a Habit record from db by id
func (hd *habitDBHD) DeleteByID(db *gorm.DB, id uint64) response.SError {
	err := db.Where("id=?", id).Delete(&Habit{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete habit fail")
	}
	return nil
}
