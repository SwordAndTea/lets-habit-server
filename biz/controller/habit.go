package controller

import (
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
	"gorm.io/gorm"
)

type HabitCtrl struct{}

// DetailedHabit a struct to represent a habit and its group user
type DetailedHabit struct {
	Habit     *dal.Habit `json:"habit"`
	UserGroup []*dal.User
}

// AddHabit add a habit and its group user info
func (c *HabitCtrl) AddHabit(h *dal.Habit, uids []dal.UID) (*DetailedHabit, response.SError) {
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

		sErr = dal.HabitDBHD.Add(tx, h)
		if sErr != nil {
			return sErr
		}
		hgs := make([]*dal.HabitGroup, 0, len(uids))
		for _, uid := range uids {
			hgs = append(hgs, &dal.HabitGroup{
				HabitID: h.ID,
				UID:     uid,
			})
		}
		sErr = dal.HabitGroupDBHD.AddMulti(tx, hgs)
		if sErr != nil {
			return sErr
		}
		return nil
	})
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{Habit: h, UserGroup: users}, nil
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
	inGroup := false
	uids := make([]dal.UID, 0, len(hgs))
	for _, hg := range hgs {
		if hg.UID == uid {
			inGroup = true
		}
		uids = append(uids, hg.UID)
	}
	if habit.PublicLevel == dal.HabitPublicLevelPrivate && !inGroup {
		return nil, response.ErrorCode_UserNoPermission.New("habit is private")
	}

	users, sErr := dal.UserDBHD.ListByUIDs(db, uids)
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{Habit: habit, UserGroup: users}, nil
}

// ListHabitsByUID func get all the habit the user joined
func (c *HabitCtrl) ListHabitsByUID(uid dal.UID) ([]*DetailedHabit, response.SError) {
	db := service.GetDBExecutor()

	// get user joined habits
	hs, sErr := dal.HabitDBHD.ListUserJoinedHabits(db, uid)
	if sErr != nil {
		return nil, sErr
	}
	hsIDs := make([]uint64, 0, len(hs))
	for _, h := range hs {
		hsIDs = append(hsIDs, h.ID)
	}

	// get habit group info for all the habits above
	hgs, sErr := dal.HabitGroupDBHD.ListByHabitIDs(db, hsIDs)
	if sErr != nil {
		return nil, sErr
	}

	// get distinct uid and get the relation of habit with its all joined user
	uidMap := make(map[dal.UID]struct{})
	uidList := make([]dal.UID, 0, len(hgs))
	habitUserMap := make(map[uint64][]dal.UID)
	var exit bool
	for _, hg := range hgs {
		_, exit = uidMap[hg.UID]
		if !exit {
			uidMap[hg.UID] = struct{}{}
			uidList = append(uidList, hg.UID)
		}
		habitUserList, ok := habitUserMap[hg.HabitID]
		if !ok {
			habitUserList = make([]dal.UID, 0, 8)
		}
		habitUserList = append(habitUserList, hg.UID)
		habitUserMap[hg.HabitID] = habitUserList
	}

	// get users and map users by its uid
	users, sErr := dal.UserDBHD.ListByUIDs(db, uidList)
	if sErr != nil {
		return nil, sErr
	}
	userIDUserMap := make(map[dal.UID]*dal.User)
	for _, u := range users {
		userIDUserMap[u.UID] = u
	}

	// construct return info
	detailedHabits := make([]*DetailedHabit, 0, len(hs))
	for _, h := range hs {
		habitUserList := habitUserMap[h.ID]
		habitUsers := make([]*dal.User, 0, len(habitUserList))
		for _, u := range habitUserList {
			habitUsers = append(habitUsers, userIDUserMap[u])
		}
		detailedHabits = append(detailedHabits, &DetailedHabit{
			Habit:     h,
			UserGroup: habitUsers,
		})
	}
	return detailedHabits, nil
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
