package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
)

type UserHabitConfig struct {
	UID                 UID    `json:"uid"`
	HabitID             uint64 `json:"habit_id"`
	CurrentStreak       uint32 `json:"current_streak"`
	LongestStreak       uint32 `json:"longest_streak"`
	RemainRecheckChance uint8  `json:"remain_recheck_chance"`
	HeatmapColor        string `json:"heatmap_color"`
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
	CurrentStreak *uint32
	LongestStreak *uint32
	HeatmapColor  *string
}

func (hd *userHabitConfigDBHD) Update(db *gorm.DB, uid UID, habitID uint64, updateFields *UserHabitConfigUpdatableFields) response.SError {
	updates := map[string]interface{}{}
	if updateFields.CurrentStreak != nil {
		updates["current_streak"] = *updateFields.CurrentStreak
	}
	if updateFields.LongestStreak != nil {
		updates["longest_streak"] = *updateFields.LongestStreak
	}
	if updateFields.HeatmapColor != nil {
		updates["longest_steak"] = *updateFields.HeatmapColor
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

func (hd *userHabitConfigDBHD) ListUserHabitConfig(db *gorm.DB, uid UID, habits []uint64) ([]*UserHabitConfig, response.SError) {
	var uhcs []*UserHabitConfig
	err := db.Model(&UserHabitConfig{}).Where("uid = ? and habit_id in (?)", uid, habits).Find(&uhcs).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list user habit config fail")
	}
	return uhcs, nil
}
