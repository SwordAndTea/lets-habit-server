package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

type UserHabitConfig struct {
	UID                     UID        `json:"uid"`
	HabitID                 uint64     `json:"habit_id"`
	CurrentStreak           uint32     `json:"current_streak"`
	LongestStreak           uint32     `json:"longest_streak"`
	StreakUpdateAt          *time.Time `json:"-"`
	RemainRetroactiveChance uint8      `json:"remain_retroactive_chance"`
	HeatmapColor            string     `json:"heatmap_color"`
}

type userHabitConfigDBHD struct{}

var UserHabitConfigDBHD = &userHabitConfigDBHD{}

func (hd *userHabitConfigDBHD) Add(db *gorm.DB, c *UserHabitConfig) response.SError {
	err := db.Create(c).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one user habit config fail")
	}
	return nil
}

type UserHabitConfigUpdatableFields struct {
	CurrentStreak  *uint32
	LongestStreak  *uint32
	StreakUpdateAt *time.Time
	HeatmapColor   *string
}

func (hd *userHabitConfigDBHD) Update(db *gorm.DB, uid UID, habitID uint64, updateFields *UserHabitConfigUpdatableFields) response.SError {
	updates := map[string]interface{}{}
	if updateFields.CurrentStreak != nil {
		updates["current_streak"] = *updateFields.CurrentStreak
	}
	if updateFields.LongestStreak != nil {
		updates["longest_streak"] = *updateFields.LongestStreak
	}
	if updateFields.StreakUpdateAt != nil {
		updates["streak_update_at"] = *updateFields.StreakUpdateAt
	}
	if updateFields.HeatmapColor != nil {
		updates["heatmap_color"] = *updateFields.HeatmapColor
	}

	if len(updates) == 0 {
		return nil
	}
	err := db.Model(&UserHabitConfig{}).Where("uid=? and habit_id=?", uid, habitID).Updates(updates).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update user habit config fail")
	}
	return nil
}

func (hd *userHabitConfigDBHD) UpdateMany(db *gorm.DB, uid UID, habitIDs []uint64, updateFields *UserHabitConfigUpdatableFields) response.SError {
	updates := map[string]interface{}{}
	if updateFields.CurrentStreak != nil {
		updates["current_streak"] = *updateFields.CurrentStreak
	}
	if updateFields.LongestStreak != nil {
		updates["longest_streak"] = *updateFields.LongestStreak
	}
	if updateFields.StreakUpdateAt != nil {
		updates["streak_update_at"] = *updateFields.StreakUpdateAt
	}
	if updateFields.HeatmapColor != nil {
		updates["heatmap_color"] = *updateFields.HeatmapColor
	}

	if len(updates) == 0 {
		return nil
	}
	err := db.Model(&UserHabitConfig{}).Where("uid=? and habit_id in (?)", uid, habitIDs).Updates(updates).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update user habit config fail")
	}
	return nil
}

func (hd *userHabitConfigDBHD) IncreaseCurrentStreakByOne(db *gorm.DB, uids []UID, habitID uint64) response.SError {
	err := db.Model(&UserHabitConfig{}).Where("uid in (?) and habit_id=?", uids, habitID).
		UpdateColumn("current_streak", gorm.Expr("current_streak + ?", 1)).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "increase current streak fail")
	}
	err = db.Model(&UserHabitConfig{}).Where("uid in (?) and habit_id=? and current_streak > longest_streak", uids, habitID).
		UpdateColumn("longest_streak", "current_streak").UpdateColumn("streak_update_at", time.Now().UTC()).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update longest streak fail")
	}
	return nil
}

func (hd *userHabitConfigDBHD) GetByUIDAndHabitID(db *gorm.DB, uid UID, habitID uint64) (*UserHabitConfig, response.SError) {
	var uhcs *UserHabitConfig
	err := db.Where("uid=? and habit_id=?", uid, habitID).First(&uhcs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get user habit config fail")
	}
	return uhcs, nil
}

func (hd *userHabitConfigDBHD) ListUserHabitConfig(db *gorm.DB, uid UID, habits []uint64) ([]*UserHabitConfig, response.SError) {
	var uhcs []*UserHabitConfig
	err := db.Where("uid=? and habit_id in (?)", uid, habits).Find(&uhcs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list user habit config fail")
	}
	return uhcs, nil
}

func (hd *userHabitConfigDBHD) DeleteByHabitIDAndUID(db *gorm.DB, habitID uint64, uid UID) response.SError {
	err := db.Where("habit_id=? and uid=?", habitID, uid).Delete(&UserHabitConfig{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete user habit config fail")
	}
	return nil
}
