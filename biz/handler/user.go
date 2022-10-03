package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/swordandtea/fhwh/biz/controller"
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
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

	sErr = r.Ctrl.EmailRegister(req.Email, dal.NewRawPassword(req.Password))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
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
