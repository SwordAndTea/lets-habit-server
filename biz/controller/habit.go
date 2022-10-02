package controller

import (
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
	"gorm.io/gorm"
)

type HabitCtrl struct{}

type DetailedHabit struct {
	Habit     *dal.Habit `json:"habit"`
	UserGroup []*dal.User
}

func (c *HabitCtrl) AddHabit(h *dal.Habit, uids []string) (*DetailedHabit, response.SError) {
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

func (c *HabitCtrl) GetHabitByID(id uint64) (*DetailedHabit, response.SError) {
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
	uids := make([]string, 0, len(hgs))
	for _, hg := range hgs {
		uids = append(uids, hg.UID)
	}

	users, sErr := dal.UserDBHD.ListByUIDs(db, uids)
	if sErr != nil {
		return nil, sErr
	}

	return &DetailedHabit{Habit: habit, UserGroup: users}, nil
}

func (c *HabitCtrl) DeleteHabitByID(id uint64, uid string) response.SError {
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
