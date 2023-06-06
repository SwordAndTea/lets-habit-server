package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/swordandtea/lets-habit-server/biz/controller"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
)

type UserRouter struct {
	Ctrl *controller.UserCtrl
}

func NewUserRouter() *UserRouter {
	return &UserRouter{Ctrl: &controller.UserCtrl{}}
}

/*********************** User Router User Auth Check ***********************/

type GetUserInfoByAuthResponse struct {
	User *dal.User `json:"user"`
}

func (r *UserRouter) GetUserInfoByAuth(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)
	uid := rc.GetString(UIDKey)
	user, sErr := r.Ctrl.GetUserByUID(dal.UID(uid))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	sErr = SetUserTokenCookie(rc, user.UID)
	if sErr != nil {
		resp.SetError(sErr)
	}
	resp.SetSuccessData(&GetUserInfoByAuthResponse{User: user})
}

/*********************** User Router User Register By Email Handler ***********************/

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *UserRegisterRequest) validate() response.SError {
	sErr := ValidateEmail(r.Email)
	if sErr != nil {
		return sErr
	}
	return nil
}

type UserRegisterResponse struct {
	User *dal.User `json:"user"`
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

	user, sErr := r.Ctrl.EmailRegister(req.Email, dal.NewRawPassword(req.Password))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	sErr = SetUserTokenCookie(rc, user.UID)
	if sErr != nil {
		resp.SetError(sErr)
	}
	resp.SetSuccessData(&UserRegisterResponse{
		User: user,
	})
}

/*********************** User Router User Register Resend Activate Email Handler ***********************/

func (r *UserRouter) ResendActivateEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	uid := rc.GetString(UIDKey)

	sErr := r.Ctrl.ResendActivateEmail(dal.UID(uid))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}

/*********************** User Router User Register Activate Email Handler ***********************/

type EmailActivateRequest struct {
	ActivateCode string `json:"activate_code"`
}

func (r *EmailActivateRequest) validate() response.SError {
	if r.ActivateCode == "" {
		return response.ErrorCode_InvalidParam.New("empty activate code")
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

	user, sErr := r.Ctrl.EmailActivate(req.ActivateCode)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	sErr = SetUserTokenCookie(rc, user.UID)
	if sErr != nil {
		resp.SetError(sErr)
	}
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
	BindCode string `json:"bind_code"`
}

func (r *ConfirmBindEmailRequest) validate() response.SError {
	if r.BindCode == "" {
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

	sErr = r.Ctrl.ConfirmBindEmail(req.BindCode)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
}

/*********************** User Router User Register By Email Handler ***********************/

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *UserLoginRequest) validate() response.SError {
	if r.Email == "" {
		return response.ErrorCode_InvalidParam.New("empty email")
	}
	if r.Password == "" {
		return response.ErrorCode_InvalidParam.New("invalid password")
	}
	return nil
}

type UserLoginResponse struct {
	User *dal.User `json:"user"`
}

func (r *UserRouter) LoginByEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &UserLoginRequest{}
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

	user, sErr := r.Ctrl.LoginByEmail(req.Email, dal.NewRawPassword(req.Password))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	sErr = SetUserTokenCookie(rc, user.UID)
	if sErr != nil {
		resp.SetError(sErr)
	}
	resp.SetSuccessData(&UserLoginResponse{
		User: user,
	})
}

/*********************** User Router Update User Base Info Handler ***********************/

type UpdateUserBaseInfoRequest struct {
	Name     string `json:"name"`
	Portrait string `json:"portrait"`
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

	portraitData, sErr := Base64ImgDecode(req.Portrait)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	uid := rc.GetString(UIDKey)
	user, sErr := r.Ctrl.UpdateUserBaseInfo(dal.UID(uid), &controller.UpdateUserBaseInfoFields{
		Name:     req.Name,
		Portrait: portraitData,
	})
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&UpdateUserBaseInfoResponse{User: user})
}

/*********************** User Router Search User Handler ***********************/

type UserSearchRequest struct {
	NameOrUID string `json:"name_or_uid"`
}

func (r *UserSearchRequest) validate() response.SError {
	if r.NameOrUID == "" {
		return response.ErrorCode_InvalidParam.New("empty search text")
	}
	return nil
}

type UserSearchResponse struct {
	Users []*controller.SimplifiedUser `json:"users"`
}

func (r *UserRouter) UserSearch(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &UserSearchRequest{}
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

	users, sErr := r.Ctrl.SearchUserByNameOrUID(req.NameOrUID)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	resp.SetSuccessData(&UserSearchResponse{Users: users})
}

/*********************** User Router Delete Account Handler ***********************/

func (r *UserRouter) DeleteAccount(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	uid := rc.GetString(UIDKey)
	sErr := r.Ctrl.DeleteAccount(dal.UID(uid))
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	ClearUserTokenCookie(rc)
}
