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
	Habit           *dal.Habit            `json:"habit"`
	UserHabitConfig *dal.UserHabitConfig  `json:"user_custom_config"`
	UserGroup       []*dal.User           `json:"user_group"`
	LogRecords      []*dal.HabitLogRecord `json:"log_records"`
	TodayLogged     bool                  `json:"today_logged"`
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
		hgs = append(hgs, &dal.HabitGroup{
			HabitID: habit.ID,
			UID:     creator,
		})
		for _, uid := range cooperators {
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
			UID:                     creator,
			HabitID:                 habit.ID,
			CurrentStreak:           0,
			LongestStreak:           0,
			RemainRetroactiveChance: 0,
			HeatmapColor:            customConfig.HeatmapColor,
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

func (c *HabitCtrl) UpdateHabit(uid dal.UID, habitID uint64, cooperatorsToAdd []dal.UID, cooperatorsToRemove []dal.UID) response.SError {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
	if sErr != nil {
		return sErr
	}
	if habit == nil {
		return response.ErrorCode_InvalidParam.New("invalid habit id, not found")
	}
	if habit.Owner != uid {
		return response.ErrorCode_UserNoPermission.New("only habit creator can update")
	}

	hgs, sErr := dal.HabitGroupDBHD.ListByHabitID(db, habitID)
	if sErr != nil {
		return sErr
	}

	// TODO: verify cooperators to Add and cooperators to delete

	if len(hgs)+len(cooperatorsToAdd)-len(cooperatorsToAdd) > CooperatorLimit {
		return response.ErrorCode_InvalidParam.New("cooperator exceed limit")
	}

	hgsToAdd := make([]*dal.HabitGroup, 0, len(cooperatorsToAdd))
	for _, cooperator := range cooperatorsToAdd {
		hgsToAdd = append(hgsToAdd, &dal.HabitGroup{
			HabitID: habitID,
			UID:     cooperator,
		})
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		sErr = dal.HabitGroupDBHD.AddMulti(tx, hgsToAdd)
		if sErr != nil {
			return sErr
		}
		sErr = dal.HabitGroupDBHD.DeleteByHabitIDAndUIDs(tx, habitID, cooperatorsToRemove)
		if sErr != nil {
			return sErr
		}
		return nil
	})
	if sErr != nil {
		return sErr
	}
	return nil
}

func (c *HabitCtrl) UpdateUserHabitConfig(uid dal.UID, habitID uint64, heatmapColor string) response.SError {
	db := service.GetDBExecutor()
	hg, sErr := dal.HabitGroupDBHD.GetByHabitIDAndUID(db, habitID, uid)
	if sErr != nil {
		return sErr
	}
	if hg == nil {
		return response.ErrorCode_InvalidParam.New("current user is not in this habit")
	}
	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		return dal.UserHabitConfigDBHD.Update(tx, uid, habitID, &dal.UserHabitConfigUpdatableFields{
			HeatmapColor: &heatmapColor,
		})
	})
	if sErr != nil {
		return sErr
	}
	return nil
}

// GetHabitByID get a habit and its group info by habit id,
// if current user not in its group, return error
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

	// TODO: get log record

	users, sErr := dal.UserDBHD.ListByUIDs(db, uids)
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{Habit: habit, UserGroup: users}, nil
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

	// get user habit log record
	habitLogRecords, sErr := dal.HabitLogRecordDBHD.ListByUIDHabitIDs(db, uid, hsIDs, fromTime, toTime)
	if sErr != nil {
		return nil, 0, sErr
	}
	habitLatestCheckTime := make(map[uint64]*time.Time)
	habitIDHLRMap := make(map[uint64][]*dal.HabitLogRecord)
	for _, hlr := range habitLogRecords {
		recordList, ok := habitIDHLRMap[hlr.HabitID]
		if !ok {
			recordList = make([]*dal.HabitLogRecord, 0, 32)
		}
		recordList = append(recordList, hlr)
		habitIDHLRMap[hlr.HabitID] = recordList
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
	if toTime.Unix() >= todayBegin.Unix() && toTime.Unix() < todayEnd.Unix() { // TODay included
		unconfirmedHabitLogRecords, sErr = dal.UnconfirmedHabitLogRecordDBHD.ListByUIDHabitIDs(db, uid, hsIDs, &todayBegin, &now)
		if sErr != nil {
			return nil, 0, sErr
		}
	}

	habitIDUHLRMap := make(map[uint64]*dal.HabitLogRecord)
	for _, uhlr := range unconfirmedHabitLogRecords {
		habitIDUHLRMap[uhlr.HabitID] = uhlr
	}

	// construct return info
	habitToClearStreak := make([]uint64, 0, len(hs))
	detailedHabits := make([]*DetailedHabit, 0, len(hs))
	for _, h := range hs {
		_, ok := habitIDUHLRMap[h.ID]
		uhc := habitIDUserHabitConfigMap[h.ID]
		detailedHabits = append(detailedHabits, &DetailedHabit{
			Habit:           h,
			UserHabitConfig: habitIDUserHabitConfigMap[h.ID],
			LogRecords:      habitIDHLRMap[h.ID],
			TodayLogged:     ok,
		})
		// check whether last log time logged
		if uhc.StreakUpdateAt != nil && uhc.StreakUpdateAt.In(toTime.Location()).Before(todayBegin) {
			latestCheckTime := habitLatestCheckTime[h.ID]
			latestCheckTimeWeekday := latestCheckTime.In(toTime.Location()).Weekday()
			// get last day need to log habit
			curWeekday := now.Weekday()
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

func (c *HabitCtrl) LogHabit(uid dal.UID, habitID uint64, logTime *time.Time) (bool /*all user checked*/, response.SError) {
	db := service.GetDBExecutor()
	habit, sErr := dal.HabitDBHD.GetByID(db, habitID)
	if sErr != nil {
		return false, sErr
	}
	if habit == nil {
		return false, response.ErrorCode_InvalidParam.New("habit not exist")
	}

	hgs, sErr := dal.HabitGroupDBHD.ListByHabitID(db, habitID)
	if sErr != nil {
		return false, sErr
	}

	inGroup := false
	for _, hg := range hgs {
		if hg.UID == uid {
			inGroup = true
			break
		}
	}
	if !inGroup {
		return false, response.ErrorCode_UserNoPermission.New("current user has not joined this habit")
	}

	now := time.Now().In(logTime.Location())
	dayBitMask := 1 << now.Add(-time.Hour*time.Duration(dal.HabitLogDelayHours)).Weekday()
	if !habit.LogDays.Has(dal.CheckDay(dayBitMask)) {
		return false, response.ErrorCode_InvalidParam.New("current day no need to log")
	}

	todayBegin, todayEnd := getTodayBeginEndTime(&now)
	allChecked := true
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
		newRecord := &dal.HabitLogRecord{
			HabitID: habitID,
			UID:     uid,
			LogAt:   now.UTC(),
		}
		logRecords = append(logRecords, newRecord)

		for _, hg := range hgs {
			if !logMap[hg.UID] {
				allChecked = false
				break
			}
		}

		if allChecked {
			sErr = dal.HabitLogRecordDBHD.AddMulti(tx, logRecords)
			if sErr != nil {
				return sErr
			}
			sErr = dal.UnconfirmedHabitLogRecordDBHD.DeleteByHabitID(tx, habitID, &todayBegin, &todayEnd)
			if sErr != nil {
				return sErr
			}
			sErr = dal.UserHabitConfigDBHD.IncreaseCurrentStreakByOne(tx, uidList, habitID)
			if sErr != nil {
				return sErr
			}
		} else {
			sErr = dal.UnconfirmedHabitLogRecordDBHD.Add(tx, newRecord)
			if sErr != nil {
				return sErr
			}
		}
		return nil
	})

	if sErr != nil {
		return false, sErr
	}

	return allChecked, nil
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
func (c *HabitCtrl) DeleteHabitByID(habitID uint64, uid dal.UID) response.SError {
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

	if habit.Owner == uid {
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
				sErr = dal.HabitDBHD.UpdateHabit(tx, habitID, &dal.HabitUpdatableFields{Owner: &uid})
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
