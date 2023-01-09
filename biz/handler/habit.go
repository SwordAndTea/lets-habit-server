package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/swordandtea/lets-habit-server/biz/controller"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"time"
)

type HabitRouter struct {
	Ctrl *controller.HabitCtrl
}

func NewHabitRouter() *HabitRouter {
	return &HabitRouter{Ctrl: &controller.HabitCtrl{}}
}

/*********************** Habit Router Create Habit Handler ***********************/

type CreateHabitRequest struct {
	Name               string                      `json:"name"`
	PublicLevel        dal.HabitPublicLevel        `json:"public_level"`
	CheckType          dal.HabitCheckType          `json:"check_type"`
	CheckFrequency     dal.HabitCheckFrequency     `json:"check_frequency"`
	CheckDeadlineDelay dal.HabitCheckDeadlineDelay `json:"check_deadline_delay"`
	Cooperators        []dal.UID                   `json:"cooperators"`
}

func (r *CreateHabitRequest) validate() response.SError {
	if r.Name == "" {
		return response.ErrorCode_InvalidParam.New("invalid name")
	}
	if !r.PublicLevel.IsValid() {
		return response.ErrorCode_InvalidParam.New("invalid public level")
	}
	if !r.CheckType.IsValid() {
		return response.ErrorCode_InvalidParam.New("invalid check type")
	}
	if !r.CheckFrequency.IsValid() {
		return response.ErrorCode_InvalidParam.New("invalid check frequency")
	}
	if !r.CheckDeadlineDelay.IsValid() {
		return response.ErrorCode_InvalidParam.New("invalid check deadline delay")
	}

	return nil
}

type CreateHabitResponse struct {
	Habit *controller.DetailedHabit `json:"habit"`
}

func (r *HabitRouter) CreateHabit(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &CreateHabitRequest{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	sErr := req.validate()
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	uid := rc.GetString(UIDKey)
	habit := &dal.Habit{
		Name:               req.Name,
		PublicLevel:        req.PublicLevel,
		CheckType:          req.CheckType,
		CheckFrequency:     req.CheckFrequency,
		CheckDeadlineDelay: req.CheckDeadlineDelay,
	}

	detailHabits, sErr := r.Ctrl.AddHabit(habit, dal.UID(uid), req.Cooperators)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&CreateHabitResponse{Habit: detailHabits})
}

/*********************** Habit Router List Habits Handler ***********************/

type ListHabitsRequest struct {
	Page     uint `query:"page"`
	PageSize uint `query:"page_size"`
}

func (r *ListHabitsRequest) validate() response.SError {
	if r.Page == 0 {
		return response.ErrorCode_InvalidParam.New("page mast greater than 0")
	}
	if r.PageSize == 0 || r.PageSize > 100 {
		return response.ErrorCode_InvalidParam.New("page size must greater than 0 and less than 100")
	}
	return nil
}

type ListHabitsResponse struct {
	Habits []*controller.DetailedHabit `json:"habits"`
	Total  uint                        `json:"total"`
}

func (r *HabitRouter) ListHabits(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &ListHabitsRequest{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	sErr := req.validate()
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	uid := rc.GetString(UIDKey)
	habits, total, sErr := r.Ctrl.ListHabitsByUID(dal.UID(uid), &dal.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	})

	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	resp.SetSuccessData(&ListHabitsResponse{
		Habits: habits,
		Total:  total,
	})
}

/*********************** Habit Router Update Habit Handler ***********************/

type UpdateHabitReqeust struct {
	Name               string                  `json:"name"`
	PublicLevel        dal.HabitPublicLevel    `json:"public_level"`
	CheckFrequency     dal.HabitCheckFrequency `json:"check_frequency"`
	CheckDeadlineDelay time.Duration           `json:"check_deadline_delay"`
}

func (r *HabitRouter) UpdateHabit(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &UpdateHabitReqeust{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	// TODO: Implement this and register to router
}
