package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/swordandtea/fhwh/biz/controller"
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
	"mime/multipart"
)

type UserRouter struct {
	Ctrl *controller.UserCtrl
}

func NewUserRouter() *UserRouter {
	return &UserRouter{Ctrl: &controller.UserCtrl{}}
}

/*********************** User Router User Register By Email Handler ***********************/

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *UserRegisterRequest) validate() response.SError {
	sErr := ValidateEmail(r.Email)
	if sErr != nil {
		return response.ErrorCode_InvalidParam.New("invalid email")
	}
	sErr = ValidatePassword(r.Password)
	if sErr != nil {
		return sErr
	}
	return nil
}

type UserRegisterResponse struct {
	UID dal.UID `json:"uid"`
}

func (r *UserRouter) RegisterByEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &UserRegisterRequest{}
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

	uid, sErr := r.Ctrl.EmailRegister(req.Email, dal.NewRawPassword(req.Password))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&UserRegisterResponse{UID: uid})
}

/*********************** User Router User Register Activate Email Handler ***********************/

type EmailActivateRequest struct {
	Code string `query:"code"`
}

func (r *EmailActivateRequest) validate() response.SError {
	if r.Code == "" {
		return response.ErrorCode_InvalidParam.New("invalid code")
	}
	return nil
}

type EmailActivateResponse struct {
	User *dal.User `json:"user"`
}

func (r *UserRouter) ActivateEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &EmailActivateRequest{}
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

	user, userToken, sErr := r.Ctrl.EmailActivate(req.Code)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	rc.Response.Header.Set("X-User-Token", userToken)
	resp.SetSuccessData(&EmailActivateResponse{User: user})
}

/*********************** User Router User Submit Bind Email Handler ***********************/

type SubmitBindEmailRequest struct {
	Email string `json:"email"`
}

func (r *SubmitBindEmailRequest) validate() response.SError {
	return ValidateEmail(r.Email)
}

func (r *UserRouter) SubmitBindEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &SubmitBindEmailRequest{}
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
	sErr = r.Ctrl.StartEmailBinding(dal.UID(uid), req.Email)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}

/*********************** User Router User Confirm Bind Email Handler ***********************/

type ConfirmBindEmailRequest struct {
	Code string `query:"code"`
}

func (r *ConfirmBindEmailRequest) validate() response.SError {
	if r.Code == "" {
		return response.ErrorCode_InvalidParam.New("missing bind code")
	}
	return nil
}

func (r *UserRouter) ConfirmBindEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &ConfirmBindEmailRequest{}
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

	sErr = r.Ctrl.ConfirmBindEmail(req.Code)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}

/*********************** User Router Update User Base Info Handler ***********************/

type UpdateUserBaseInfoRequest struct {
	Name     string                `form:"name"`
	Portrait *multipart.FileHeader `form:"portrait"`
}

func (r *UpdateUserBaseInfoRequest) validate() response.SError {
	if r.Portrait == nil || r.Portrait.Size == 0 {
		return response.ErrorCode_InvalidParam.New("portrait file empty")
	}
	return nil
}

type UpdateUserBaseInfoResponse struct {
	User *dal.User `json:"user"`
}

func (r *UserRouter) UpdateUserBaseInfo(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &UpdateUserBaseInfoRequest{}
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
	user, sErr := r.Ctrl.UpdateUserBaseInfo(dal.UID(uid), &controller.UpdateUserBaseInfoFields{
		Name:     req.Name,
		Portrait: req.Portrait,
	})
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&UpdateUserBaseInfoResponse{User: user})
}
