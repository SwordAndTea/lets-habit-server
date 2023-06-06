package controller

import (
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"github.com/swordandtea/lets-habit-server/util"
	"gorm.io/gorm"
	"time"
)

type HabitCtrl struct{}

const CooperatorLimit = 5

type HabitCustomConfig struct {
	HeatmapColor string `json:"heatmap_color"`
}

// DetailedHabit a struct to represent a habit and its group user
type DetailedHabit struct {
	Habit               *dal.Habit            `json:"habit"`
	UserHabitConfig     *dal.UserHabitConfig  `json:"user_custom_config"`
	Cooperators         []*SimplifiedUser     `json:"cooperators"`
	CooperatorLogStatus map[dal.UID]bool      `json:"cooperator_log_status"`
	LogRecords          []*dal.HabitLogRecord `json:"log_records"`
	TodayLogged         bool                  `json:"today_logged"`
}

// AddHabit add a habit and its group user info
func (c *HabitCtrl) AddHabit(habit *dal.Habit, creator dal.UID, cooperators []dal.UID, customConfig *HabitCustomConfig) (*DetailedHabit, response.SError) {
	if len(cooperators) > CooperatorLimit {
		return nil, response.ErrorCode_InvalidParam.New("cooperator exceed limit")
	}
	db := service.GetDBExecutor()
	users, sErr := dal.UserDBHD.ListByUIDs(db, cooperators)
	if sErr != nil {
		return nil, sErr
	}
	if len(users) != len(cooperators) {
		return nil, response.ErrorCode_InvalidParam.New("has non-exist uid")
	}

	habit.Owner = creator
	habit.CreateAt = time.Now().UTC()
	sErr = WithDBTx(nil, func(tx *gorm.DB) response.SError {
		sErr = dal.HabitDBHD.Add(tx, habit)
		if sErr != nil {
			return sErr
		}
		hgs := make([]*dal.HabitGroup, 0, len(cooperators)+1)
		uhcs := make([]*dal.UserHabitConfig, 0, len(cooperators)+1)
		hgs = append(hgs, &dal.HabitGroup{
			HabitID: habit.ID,
			UID:     creator,
		})
		uhcs = append(uhcs, &dal.UserHabitConfig{
			UID:                     creator,
			HabitID:                 habit.ID,
			CurrentStreak:           0,
			LongestStreak:           0,
			RemainRetroactiveChance: 0,
			HeatmapColor:            customConfig.HeatmapColor,
		})

		for _, uid := range cooperators {
			if uid != creator {
				hgs = append(hgs, &dal.HabitGroup{
					HabitID: habit.ID,
					UID:     uid,
				})
				uhcs = append(uhcs, &dal.UserHabitConfig{
					UID:                     uid,
					HabitID:                 habit.ID,
					CurrentStreak:           0,
					LongestStreak:           0,
					RemainRetroactiveChance: 0,
					HeatmapColor:            customConfig.HeatmapColor,
				})
			}
		}
		sErr = dal.HabitGroupDBHD.AddMulti(tx, hgs)
		if sErr != nil {
			return sErr
		}

		sErr = dal.UserHabitConfigDBHD.AddMulti(tx, uhcs)
		if sErr != nil {
			return sErr
		}

		return nil
	})
	if sErr != nil {
		return nil, sErr
	}

	SimplifiedUsers := make([]*SimplifiedUser, 0, len(users))
	for _, u := range users {
		SimplifiedUsers = append(SimplifiedUsers, &SimplifiedUser{
			UID:      u.UID,
			Name:     u.Name,
			Portrait: u.PortraitURL,
		})
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
		Cooperators: SimplifiedUsers,
	}, nil
}

type HabitUpdatableInfo struct {
	Name                string    `json:"name"`
	Identity            string    `json:"identity"`
	CooperatorsToAdd    []dal.UID `json:"cooperators_to_add"`
	CooperatorsToDelete []dal.UID `json:"cooperators_to_delete"`
}

func (u *HabitUpdatableInfo) IsValid() bool {
	return u.Name != "" || u.Identity != "" || len(u.CooperatorsToAdd) != 0 || len(u.CooperatorsToDelete) != 0
}

type UserHabitConfigUpdatableField struct {
	HeatmapColor string `json:"heatmap_color"`
}

func (u *UserHabitConfigUpdatableField) IsValid() bool {
	return u.HeatmapColor != ""
}

func (c *HabitCtrl) UpdateHabit(uid dal.UID, habitID uint64, basicInfo *HabitUpdatableInfo,
	customConfig *UserHabitConfigUpdatableField) response.SError {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
	if sErr != nil {
		return sErr
	}
	if habit == nil {
		return response.ErrorCode_InvalidParam.New("invalid habit id, not found")
	}

	habitGroups, sErr := dal.HabitGroupDBHD.ListByHabitID(db, habitID)
	if sErr != nil {
		return sErr
	}

	inGroup := false
	for _, hg := range habitGroups {
		if hg.UID == uid {
			inGroup = true
			break
		}
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		// TODO: verify cooperators to Add and cooperators to delete
		if basicInfo.IsValid() {
			if habit.Owner != uid {
				return response.ErrorCode_UserNoPermission.New("current user not own this habit")
			}
			if len(habitGroups)+len(basicInfo.CooperatorsToAdd)-len(basicInfo.CooperatorsToDelete) > CooperatorLimit {
				return response.ErrorCode_InvalidParam.New("cooperator exceed limit")
			}

			if basicInfo.Name != "" || basicInfo.Identity != "" {
				sErr = dal.HabitDBHD.UpdateHabit(tx, habitID, &dal.HabitUpdatableFields{
					Name:     basicInfo.Name,
					Identity: basicInfo.Identity,
				})
				if sErr != nil {
					return sErr
				}
			}

			if len(basicInfo.CooperatorsToAdd) != 0 {
				hgsToAdd := make([]*dal.HabitGroup, 0, len(basicInfo.CooperatorsToAdd))
				for _, cooperator := range basicInfo.CooperatorsToAdd {
					hgsToAdd = append(hgsToAdd, &dal.HabitGroup{
						HabitID: habitID,
						UID:     cooperator,
					})
				}
				sErr = dal.HabitGroupDBHD.AddMulti(tx, hgsToAdd)
				if sErr != nil {
					return sErr
				}
			}
			if len(basicInfo.CooperatorsToDelete) != 0 {
				sErr = dal.HabitGroupDBHD.DeleteByHabitIDAndUIDs(tx, habitID, basicInfo.CooperatorsToDelete)
				if sErr != nil {
					return sErr
				}
			}
		}

		if customConfig.IsValid() {
			if !inGroup {
				return response.ErrorCode_UserNoPermission.New("current user not in this habit")
			}
			sErr = dal.UserHabitConfigDBHD.Update(tx, uid, habitID, &dal.UserHabitConfigUpdatableFields{
				HeatmapColor: customConfig.HeatmapColor,
			})
			if sErr != nil {
				return sErr
			}
		}
		return nil
	})
	if sErr != nil {
		return sErr
	}
	return nil
}

// GetHabitByID get a habit and its group info by habit id,
// if current user not in its group, return error
func (c *HabitCtrl) GetHabitByID(habitID uint64, uid dal.UID) (*DetailedHabit, response.SError) {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
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
	inGroup := false
	uids := make([]dal.UID, 0, len(hgs))
	for _, hg := range hgs {
		if hg.UID == uid {
			inGroup = true
		}
		uids = append(uids, hg.UID)
	}

	if !inGroup {
		return nil, response.ErrorCode_UserNoPermission.New("current user not participated in this habit")
	}

	users, sErr := dal.UserDBHD.ListByUIDs(db, uids)
	if sErr != nil {
		return nil, sErr
	}

	userHabitConfig, sErr := dal.UserHabitConfigDBHD.GetByUIDAndHabitID(db, uid, habitID)
	if sErr != nil {
		return nil, sErr
	}

	SimplifiedUsers := make([]*SimplifiedUser, 0, len(users))
	for _, u := range users {
		SimplifiedUsers = append(SimplifiedUsers, &SimplifiedUser{
			UID:      u.UID,
			Name:     u.Name,
			Portrait: u.PortraitURL,
		})
	}

	// TODO: return log record info, check should update streak info

	return &DetailedHabit{
		Habit:           habit,
		UserHabitConfig: userHabitConfig,
		Cooperators:     SimplifiedUsers,
	}, nil
}

func getTodayBeginEndTime(now *time.Time) (time.Time, time.Time) {
	ny, nm, nd := now.Date()
	todayBegin := time.Date(ny, nm, nd, dal.HabitLogDelayHours, 0, 0, 0, now.Location())
	if now.Hour() < dal.HabitLogDelayHours {
		todayBegin = todayBegin.AddDate(0, 0, -1)
	}
	todayEnd := todayBegin.AddDate(0, 0, 1)
	return todayBegin, todayEnd
}

// ListHabitsByUID get all the habit the user joined
func (c *HabitCtrl) ListHabitsByUID(uid dal.UID, pagination *dal.Pagination, fromTime *time.Time, toTime *time.Time) ([]*DetailedHabit, uint, response.SError) {
	db := service.GetDBExecutor()

	// get user joined habits
	habits, total, sErr := dal.HabitDBHD.ListUserJoinedHabits(db, uid, pagination)
	if sErr != nil {
		return nil, 0, sErr
	}
	habitIDs := make([]uint64, 0, len(habits))
	for _, h := range habits {
		habitIDs = append(habitIDs, h.ID)
	}

	// get user habit config
	userHabitConfigs, sErr := dal.UserHabitConfigDBHD.ListUserHabitConfig(db, uid, habitIDs)
	if sErr != nil {
		return nil, 0, sErr
	}

	userHabitConfigMap := make(map[uint64]*dal.UserHabitConfig)
	for _, uhc := range userHabitConfigs {
		userHabitConfigMap[uhc.HabitID] = uhc
	}

	// get habit cooperator infos
	habitGroups, sErr := dal.HabitGroupDBHD.ListByHabitIDs(db, habitIDs)
	if sErr != nil {
		return nil, 0, sErr
	}

	distinctUserMap := make(map[dal.UID]struct{})
	for _, hg := range habitGroups {
		distinctUserMap[hg.UID] = struct{}{}
	}
	distinctUserList := make([]dal.UID, 0, len(distinctUserMap))
	for k := range distinctUserMap {
		distinctUserList = append(distinctUserList, k)
	}
	users, sErr := dal.UserDBHD.ListByUIDs(db, distinctUserList)
	if sErr != nil {
		return nil, 0, sErr
	}
	userMap := make(map[dal.UID]*SimplifiedUser)
	for _, u := range users {
		userMap[u.UID] = &SimplifiedUser{
			UID:      u.UID,
			Name:     u.Name,
			Portrait: u.PortraitURL,
		}
	}
	habitCooperatorMap := make(map[uint64][]*SimplifiedUser)
	for _, hg := range habitGroups {
		cooperators, ok := habitCooperatorMap[hg.HabitID]
		if !ok {
			cooperators = make([]*SimplifiedUser, 0, 8)
		}
		cooperators = append(cooperators, userMap[hg.UID])
		habitCooperatorMap[hg.HabitID] = cooperators
	}

	// get user habit log record
	habitLogRecords, sErr := dal.HabitLogRecordDBHD.ListByUIDHabitIDs(db, uid, habitIDs, fromTime, toTime)
	if sErr != nil {
		return nil, 0, sErr
	}
	habitLatestCheckTime := make(map[uint64]*time.Time)
	habitLogRecordMap := make(map[uint64][]*dal.HabitLogRecord)
	for _, hlr := range habitLogRecords {
		recordList, ok := habitLogRecordMap[hlr.HabitID]
		if !ok {
			recordList = make([]*dal.HabitLogRecord, 0, 32)
		}
		recordList = append(recordList, hlr)
		habitLogRecordMap[hlr.HabitID] = recordList
		latestCheckTime, ok := habitLatestCheckTime[hlr.HabitID]
		if !ok {
			habitLatestCheckTime[hlr.HabitID] = &hlr.LogAt
		} else if latestCheckTime.Before(hlr.LogAt) {
			habitLatestCheckTime[hlr.HabitID] = &hlr.LogAt
		}
	}

	now := time.Now().In(toTime.Location())
	if !toTime.Before(now) {
		toTime = &now
	}
	todayBegin, todayEnd := getTodayBeginEndTime(&now)

	var unconfirmedHabitLogRecords []*dal.HabitLogRecord
	if toTime.Unix() >= todayBegin.Unix() && toTime.Unix() < todayEnd.Unix() { // today included
		unconfirmedHabitLogRecords, sErr = dal.UnconfirmedHabitLogRecordDBHD.ListByUIDHabitIDs(db, uid, habitIDs, &todayBegin, &now)
		if sErr != nil {
			return nil, 0, sErr
		}
	}

	habitIDUHLRMap := make(map[uint64]*dal.HabitLogRecord)
	habitCooperatorLogStatus := make(map[uint64]map[dal.UID]bool)
	for _, uhlr := range unconfirmedHabitLogRecords {
		habitIDUHLRMap[uhlr.HabitID] = uhlr
		cooperatorLogStatus, ok := habitCooperatorLogStatus[uhlr.HabitID]
		if !ok {
			cooperatorLogStatus = make(map[dal.UID]bool)
		}
		cooperatorLogStatus[uhlr.UID] = true
		habitCooperatorLogStatus[uhlr.HabitID] = cooperatorLogStatus
	}

	// construct return info
	habitToClearStreak := make([]uint64, 0, len(habits))
	detailedHabits := make([]*DetailedHabit, 0, len(habits))
	startWeekday := now.Weekday()
	for _, h := range habits {
		_, ok := habitIDUHLRMap[h.ID]
		uhc := userHabitConfigMap[h.ID]
		cooperatorLogStatus, exist := habitCooperatorLogStatus[h.ID]
		if !exist {
			cooperatorLogStatus = map[dal.UID]bool{}
		}
		detailedHabits = append(detailedHabits, &DetailedHabit{
			Habit:               h,
			UserHabitConfig:     userHabitConfigMap[h.ID],
			Cooperators:         habitCooperatorMap[h.ID],
			CooperatorLogStatus: cooperatorLogStatus,
			LogRecords:          habitLogRecordMap[h.ID],
			TodayLogged:         ok,
		})
		// check whether last log time logged
		if uhc.StreakUpdateAt != nil && uhc.StreakUpdateAt.In(toTime.Location()).Before(todayBegin) {
			latestCheckTime := habitLatestCheckTime[h.ID]
			latestCheckTimeWeekday := latestCheckTime.In(toTime.Location()).Weekday()
			// get last day need to log habit
			curWeekday := startWeekday
			for i := 1; i < 7; i++ {
				targetWeekDay := int(curWeekday) - i
				if targetWeekDay < 0 {
					targetWeekDay = 7 + targetWeekDay
				}
				if h.LogDays.Has(dal.CheckDay(1 << curWeekday)) {
					if targetWeekDay != int(latestCheckTimeWeekday) {
						habitToClearStreak = append(habitToClearStreak, h.ID)
					}
					break
				}
			}
		}
	}

	if len(habitToClearStreak) != 0 {
		sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
			sErr = dal.UserHabitConfigDBHD.UpdateMany(tx, uid, habitToClearStreak, &dal.UserHabitConfigUpdatableFields{
				CurrentStreak:  util.LiteralValuePtr(uint32(0)),
				StreakUpdateAt: util.LiteralValuePtr(now.UTC()),
			})
			if sErr != nil {
				return sErr
			}
			return nil
		})
		if sErr != nil {
			return nil, 0, sErr
		}
	}
	return detailedHabits, total, nil
}

func (c *HabitCtrl) LogHabit(uid dal.UID, habitID uint64, logTime *time.Time) (*dal.HabitLogRecord, response.SError) {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
	if sErr != nil {
		return nil, sErr
	}
	if habit == nil {
		return nil, response.ErrorCode_InvalidParam.New("habit not exist")
	}

	hgs, sErr := dal.HabitGroupDBHD.ListByHabitID(db, habitID)
	if sErr != nil {
		return nil, sErr
	}

	inGroup := false
	for _, hg := range hgs {
		if hg.UID == uid {
			inGroup = true
			break
		}
	}
	if !inGroup {
		return nil, response.ErrorCode_UserNoPermission.New("current user has not joined this habit")
	}

	now := time.Now().In(logTime.Location())
	dayBitMask := 1 << now.Add(-time.Hour*time.Duration(dal.HabitLogDelayHours)).Weekday()
	if !habit.LogDays.Has(dal.CheckDay(dayBitMask)) {
		return nil, response.ErrorCode_InvalidParam.New("current day no need to log")
	}

	todayBegin, todayEnd := getTodayBeginEndTime(&now)
	allChecked := true
	newRecord := &dal.HabitLogRecord{
		HabitID: habitID,
		UID:     uid,
		LogAt:   now.UTC(),
	}
	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		logRecords, sErr := dal.UnconfirmedHabitLogRecordDBHD.ListByHabitID(tx, habitID, &todayBegin, &todayEnd)
		if sErr != nil {
			return sErr
		}

		logMap := make(map[dal.UID]bool)
		uidList := make([]dal.UID, 0, len(logRecords))
		for _, lr := range logRecords {
			if lr.UID == uid {
				return response.ErrorCode_InvalidParam.New("already logged today")
			}
			lr.ID = 0 // clear primary key id
			logMap[lr.UID] = true
			uidList = append(uidList, lr.UID)
		}
		logMap[uid] = true

		sErr = dal.UnconfirmedHabitLogRecordDBHD.Add(tx, newRecord)
		if sErr != nil {
			return sErr
		}

		for _, hg := range hgs {
			if !logMap[hg.UID] {
				allChecked = false
				break
			}
		}
		if allChecked {
			newRecord.ID = 0
			logRecords = append(logRecords, newRecord)
			sErr = dal.HabitLogRecordDBHD.AddMulti(tx, logRecords)
			if sErr != nil {
				return sErr
			}
			//sErr = dal.UnconfirmedHabitLogRecordDBHD.DeleteByHabitID(tx, habitID, &todayBegin, &todayEnd)
			//if sErr != nil {
			//	return sErr
			//}
			uidList = append(uidList, uid)
			sErr = dal.UserHabitConfigDBHD.IncreaseCurrentStreakByOne(tx, uidList, habitID)
			if sErr != nil {
				return sErr
			}
		}
		return nil
	})

	if sErr != nil {
		return nil, sErr
	}

	if allChecked {
		return newRecord, nil
	}
	return nil, nil
}

func deleteHabitCommonInfo(tx *gorm.DB, habitID uint64, uid dal.UID) response.SError {
	sErr := dal.HabitGroupDBHD.DeleteByHabitIDAndUID(tx, habitID, uid)
	if sErr != nil {
		return sErr
	}
	sErr = dal.UserHabitConfigDBHD.DeleteByHabitIDAndUID(tx, habitID, uid)
	if sErr != nil {
		return sErr
	}
	sErr = dal.HabitLogRecordDBHD.DeleteByHabitIDAndUID(tx, habitID, uid)
	if sErr != nil {
		return sErr
	}
	sErr = dal.UnconfirmedHabitLogRecordDBHD.DeleteByHabitIDAndUID(tx, habitID, uid)
	if sErr != nil {
		return sErr
	}
	return nil
}

// DeleteHabitByID delete a habit, only the owner can delete
// and all the user inside its group will be removed for their habits list
func (c *HabitCtrl) DeleteHabitByID(uid dal.UID, habitID uint64) response.SError {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
	if sErr != nil {
		return sErr
	}
	if habit == nil {
		return response.ErrorCode_InvalidParam.New("habit not exist")
	}

	hg, sErr := dal.HabitGroupDBHD.GetByHabitIDAndUID(db, habitID, uid)
	if sErr != nil {
		return sErr
	}
	if hg == nil {
		return response.ErrorCode_UserNoPermission.New("current user has not joined this habit")
	}

	if habit.Owner == uid { // is habit owner
		// try to find a successor
		var successor *dal.HabitGroup
		successor, sErr = dal.HabitGroupDBHD.GetByHabitIDAndExcludeUID(db, habitID, uid)
		if sErr != nil {
			return sErr
		}
		if successor == nil { // no successor means current use is the last one participate in this habit
			sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
				sErr = deleteHabitCommonInfo(tx, habitID, uid)
				if sErr != nil {
					return sErr
				}
				return dal.HabitDBHD.DeleteByID(db, habitID)
			})
		} else {
			sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
				sErr = dal.HabitDBHD.UpdateHabit(tx, habitID, &dal.HabitUpdatableFields{Owner: successor.UID})
				if sErr != nil {
					return sErr
				}
				return deleteHabitCommonInfo(tx, habitID, uid)
			})
		}
	} else {
		sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
			return deleteHabitCommonInfo(tx, habitID, uid)
		})
	}
	if sErr != nil {
		return sErr
	}
	return nil
}
