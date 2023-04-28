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
	Name         string                        `json:"name"`
	Identity     *string                       `json:"identity"`
	Cooperators  []dal.UID                     `json:"cooperators"`
	CheckDays    dal.CheckDay                  `json:"log_days"`
	CustomConfig *controller.HabitCustomConfig `json:"custom_config"`
}

func (r *CreateHabitRequest) validate() response.SError {
	if r.Name == "" {
		return response.ErrorCode_InvalidParam.New("invalid name")
	}

	if !r.CheckDays.IsValid() {
		return response.ErrorCode_InvalidParam.New("invalid log days")
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
		Name:     req.Name,
		Identity: req.Identity,
		LogDays:  req.CheckDays,
	}

	detailHabits, sErr := r.Ctrl.AddHabit(habit, dal.UID(uid), req.Cooperators, req.CustomConfig)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&CreateHabitResponse{Habit: detailHabits})
}

/*********************** Habit Router Get Habit Handler ***********************/

type GetHabitRequest struct {
	ID uint64 `path:"id"`
}

func (r *GetHabitRequest) validate() response.SError {
	if r.ID == 0 {
		return response.ErrorCode_InvalidParam.New("invalid habit id")
	}
	return nil
}

type GetHabitResponse struct {
	Habit *controller.DetailedHabit `json:"habit"`
}

func (r *HabitRouter) GetHabit(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &GetHabitRequest{}
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
	habit, sErr := r.Ctrl.GetHabitByID(req.ID, dal.UID(uid))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	resp.SetSuccessData(&GetHabitResponse{Habit: habit})
}

/*********************** Habit Router List Habits Handler ***********************/

type ListHabitsRequest struct {
	Page          uint   `query:"page"`
	PageSize      uint   `query:"page_size"`
	FromTimestamp string `query:"from_time"`
	ToTimestamp   string `query:"to_time"`
	FromTime      *time.Time
	ToTime        *time.Time
}

func (r *ListHabitsRequest) validate() response.SError {
	if r.Page == 0 {
		return response.ErrorCode_InvalidParam.New("page mast greater than 0")
	}
	if r.PageSize == 0 || r.PageSize > 100 {
		return response.ErrorCode_InvalidParam.New("page size must greater than 0 and less than 100")
	}
	fromTime, err := time.Parse(time.RFC3339, r.FromTimestamp)
	if err != nil {
		return response.ErrorCode_InvalidParam.New("invalid from timestamp format")
	}
	r.FromTime = &fromTime

	toTime, err := time.Parse(time.RFC3339, r.ToTimestamp)
	if err != nil {
		return response.ErrorCode_InvalidParam.New("invalid to timestamp format")
	}
	r.ToTime = &toTime

	if fromTime.Location().String() != toTime.Location().String() {
		return response.ErrorCode_InvalidParam.New("from time and to time are not in the same timezone")
	}

	if !fromTime.Before(toTime) {
		return response.ErrorCode_InvalidParam.New("from time should be earlier than to time")
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
	}, req.FromTime, req.ToTime)

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
	HabitID    uint64                                   `path:"id"`
	BasicInfo  controller.HabitUpdatableInfo            `json:"basic_info"`
	CustomInfo controller.UserHabitConfigUpdatableField `json:"custom_info"`
}

func (r *UpdateHabitReqeust) validate() response.SError {
	if r.HabitID == 0 {
		return response.ErrorCode_InvalidParam.New("invalid habit id")
	}
	if !r.BasicInfo.IsValid() && !r.CustomInfo.IsValid() {
		return response.ErrorCode_InvalidParam.New("no field need to update")
	}
	return nil
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

	sErr := req.validate()
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	uid := rc.GetString(UIDKey)

	sErr = r.Ctrl.UpdateHabit(dal.UID(uid), req.HabitID, &req.BasicInfo, &req.CustomInfo)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}

/*********************** Habit Router Log Habit Handler ***********************/

type LogHabitRequest struct {
	HabitID      uint64 `path:"id"`
	LogTimestamp string `json:"log_time"`
	LogTime      *time.Time
}

func (r *LogHabitRequest) validate() response.SError {
	if r.HabitID == 0 {
		return response.ErrorCode_InvalidParam.New("invalid habit id")
	}
	logTime, err := time.Parse(time.RFC3339, r.LogTimestamp)
	if err != nil {
		return response.ErrorCode_InvalidParam.New("invalid log timestamp format")
	}
	r.LogTime = &logTime
	return nil
}

type LogHabitResponse struct {
	LogRecord *dal.HabitLogRecord `json:"log_record"`
}

func (r *HabitRouter) LogHabit(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &LogHabitRequest{}
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
	logRecord, sErr := r.Ctrl.LogHabit(dal.UID(uid), req.HabitID, req.LogTime)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	resp.SetSuccessData(&LogHabitResponse{
		LogRecord: logRecord,
	})
}

/*********************** Habit Router Delete Habit Handler ***********************/

type DeleteHabitRequest struct {
	HabitID uint64 `path:"id"`
}

func (r *HabitRouter) DeleteHabit(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &DeleteHabitRequest{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	uid := rc.GetString(UIDKey)
	sErr := r.Ctrl.DeleteHabitByID(dal.UID(uid), req.HabitID)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}
