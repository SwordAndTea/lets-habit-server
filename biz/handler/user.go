package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/swordandtea/lets-habit-server/biz/controller"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/nullable"
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
		return sErr
	}
	sErr = ValidatePassword(r.Password)
	if sErr != nil {
		return sErr
	}
	return nil
}

type UserRegisterResponse struct {
	UID       dal.UID `json:"uid"`
	PollToken string  `json:"poll_token"`
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
	pollToken, sErr := GeneratePollToken(uid)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}
	resp.SetSuccessData(&UserRegisterResponse{
		UID:       uid,
		PollToken: pollToken,
	})
}

/*********************** User Router User Register Activate Email Check Handler ***********************/

type EmailActivatedCheckRequest struct {
	PollToken string `query:"poll_token"`
}

type EmailActivateCheckResponse struct {
	Activated bool `json:"activated"`
}

func (r *UserRouter) CheckEmailActivated(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &EmailActivatedCheckRequest{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	uid, sErr := ExtractPollToken(req.PollToken)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	activated, sErr := r.Ctrl.CheckEmailActivated(uid)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	resp.SetSuccessData(&EmailActivateCheckResponse{Activated: activated})
}

/*********************** User Router User Register Resend Activate Email Handler ***********************/

type ResendActivateEmailRequest struct {
	PollToken string `json:"poll_token"`
}

func (r *UserRouter) ResendActivateEmail(ctx context.Context, rc *app.RequestContext) {
	resp := response.NewHTTPResponse(rc)
	defer resp.ReturnWithLog(ctx, rc)

	req := &ResendActivateEmailRequest{}
	err := rc.BindAndValidate(req)
	if err != nil {
		resp.SetError(BindAndValidateErr(err))
		return
	}

	uid, sErr := ExtractPollToken(req.PollToken)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	sErr = r.Ctrl.ResendActivateEmail(uid)
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

	user, userToken, sErr := r.Ctrl.EmailActivate(req.ActivateCode)
	if sErr != nil {
		resp.SetError(sErr)
		return
	}

	rc.Response.Header.Set(UserTokenHeader, userToken)
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
	User      *dal.User           `json:"user"`
	PollToken nullable.NullString `json:"poll_token"`
	UserToken nullable.NullString `json:"user_token"`
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

	pollToken := nullable.NullString{}
	userToken := nullable.NullString{}
	if user.UserRegisterType == dal.UserRegisterTypeEmail && !user.EmailActive.Get() {
		pollTokenStr, sErr := GeneratePollToken(user.UID)
		if sErr != nil {
			resp.SetError(sErr)
			return
		}
		pollToken.Set(pollTokenStr)
	} else {
		userTokenStr, sErr := GenerateUserToken(user.UID)
		if sErr != nil {
			resp.SetError(sErr)
			return
		}
		userToken.Set(userTokenStr)
	}
	resp.SetSuccessData(&UserLoginResponse{
		User:      user,
		PollToken: pollToken,
		UserToken: userToken,
	})
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
