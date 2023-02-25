package controller

import (
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"gorm.io/gorm"
	"time"
)

type HabitCtrl struct{}

type HabitCustomConfig struct {
	HeatmapColor string `json:"heatmap_color"`
}

// DetailedHabit a struct to represent a habit and its group user
type DetailedHabit struct {
	Habit           *dal.Habit              `json:"habit"`
	UserHabitConfig *dal.UserHabitConfig    `json:"user_custom_config"`
	UserGroup       []*dal.User             `json:"user_group"`
	CheckRecords    []*dal.HabitCheckRecord `json:"check_records"`
}

// AddHabit add a habit and its group user info
func (c *HabitCtrl) AddHabit(habit *dal.Habit, creator dal.UID, uids []dal.UID, customConfig *HabitCustomConfig) (*DetailedHabit, response.SError) {
	var sErr response.SError
	var users []*dal.User
	sErr = WithDBTx(func(tx *gorm.DB) response.SError {
		users, sErr = dal.UserDBHD.ListByUIDs(tx, uids)
		if sErr != nil {
			return nil
		}
		if len(users) != len(uids) {
			return response.ErrorCode_InvalidParam.New("has non-exist uid")
		}

		habit.Creator = creator
		habit.CreateAt = time.Now().UTC()
		sErr = dal.HabitDBHD.Add(tx, habit)
		if sErr != nil {
			return sErr
		}
		hgs := make([]*dal.HabitGroup, 0, len(uids)+1)
		hgs = append(hgs, &dal.HabitGroup{
			HabitID: habit.ID,
			UID:     creator,
		})
		for _, uid := range uids {
			if uid != creator {
				hgs = append(hgs, &dal.HabitGroup{
					HabitID: habit.ID,
					UID:     uid,
				})
			}
		}
		sErr = dal.HabitGroupDBHD.AddMulti(tx, hgs)
		if sErr != nil {
			return sErr
		}

		uhc := &dal.UserHabitConfig{
			UID:                 creator,
			HabitID:             habit.ID,
			CurrentStreak:       0,
			LongestStreak:       0,
			RemainRecheckChance: 0,
			HeatmapColor:        customConfig.HeatmapColor,
		}

		sErr = dal.UserHabitConfigDBHD.Add(tx, uhc)
		if sErr != nil {
			return sErr
		}

		return nil
	})
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{
		Habit: habit,
		UserHabitConfig: &dal.UserHabitConfig{
			UID:           creator,
			HabitID:       habit.ID,
			CurrentStreak: 0,
			LongestStreak: 0,
			HeatmapColor:  customConfig.HeatmapColor,
		},
		UserGroup: users,
	}, nil
}

// GetHabitByID get a habit and its group info by habit id,
// if habit is private and current user not in its group, return error
func (c *HabitCtrl) GetHabitByID(id uint64, uid dal.UID) (*DetailedHabit, response.SError) {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, id)
	if sErr != nil {
		return nil, sErr
	}
	if habit == nil {
		return nil, response.ErrorCode_InvalidParam.New("habit id not exist")
	}

	hgs, sErr := dal.HabitGroupDBHD.ListByHabitID(db, habit.ID)
	if sErr != nil {
		return nil, sErr
	}

	// check current is in group or not
	uids := make([]dal.UID, 0, len(hgs))
	for _, hg := range hgs {
		uids = append(uids, hg.UID)
	}

	users, sErr := dal.UserDBHD.ListByUIDs(db, uids)
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{Habit: habit, UserGroup: users}, nil
}

// ListHabitsByUID get all the habit the user joined
func (c *HabitCtrl) ListHabitsByUID(uid dal.UID, pagination *dal.Pagination, fromTime *time.Time, toTime *time.Time) ([]*DetailedHabit, uint, response.SError) {
	db := service.GetDBExecutor()

	// get user joined habits
	hs, total, sErr := dal.HabitDBHD.ListUserJoinedHabits(db, uid, pagination)
	if sErr != nil {
		return nil, 0, sErr
	}
	hsIDs := make([]uint64, 0, len(hs))
	for _, h := range hs {
		hsIDs = append(hsIDs, h.ID)
	}

	// get user habit config
	uhcs, sErr := dal.UserHabitConfigDBHD.ListUserHabitConfig(db, uid, hsIDs)
	if sErr != nil {
		return nil, 0, sErr
	}

	habitIDUserHabitConfigMap := make(map[uint64]*dal.UserHabitConfig)
	for _, uhc := range uhcs {
		habitIDUserHabitConfigMap[uhc.HabitID] = uhc
	}

	// get user habit check record
	hcrs, sErr := dal.HabitCheckRecordDBHD.ListByUIDHabitIDs(db, uid, hsIDs, fromTime, toTime)
	if sErr != nil {
		return nil, 0, sErr
	}
	habitIDHCRMap := make(map[uint64][]*dal.HabitCheckRecord)
	for _, hcr := range hcrs {
		hcrList, ok := habitIDHCRMap[hcr.HabitID]
		if !ok {
			hcrList = make([]*dal.HabitCheckRecord, 0, 32)
		}
		hcrList = append(hcrList, hcr)
		habitIDHCRMap[hcr.HabitID] = hcrList
	}

	// construct return info
	detailedHabits := make([]*DetailedHabit, 0, len(hs))
	for _, h := range hs {
		detailedHabits = append(detailedHabits, &DetailedHabit{
			Habit:           h,
			UserHabitConfig: habitIDUserHabitConfigMap[h.ID],
			CheckRecords:    habitIDHCRMap[h.ID],
		})
	}
	return detailedHabits, total, nil
}

// DeleteHabitByID delete a habit, only the owner can delete
// and all the user inside its group will be removed for their habits list
func (c *HabitCtrl) DeleteHabitByID(id uint64, uid dal.UID) response.SError {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, id)
	if sErr != nil {
		return sErr
	}

	if habit.Creator != uid {
		return response.ErrorCode_UserNoPermission.New("format")
	}

	return dal.HabitDBHD.DeleteByID(db, id)
}
