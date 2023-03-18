// Code generated by hertz generator.

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/swordandtea/lets-habit-server/biz/handler"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)

	// your code ...
	apiV1 := r.Group("api/v1")

	// register user related api
	userRouter := handler.NewUserRouter()
	{
		apiV1.GET("/user", handler.UserTokenVerify(), userRouter.GetUserInfoByAuth)
		apiV1.GET("/user/ping", handler.UserTokenVerify())
		apiV1.POST("/user/register/email", userRouter.RegisterByEmail)
		//apiV1.GET("/user/register/email/activate/check", userRouter.CheckEmailActivated)
		apiV1.POST("/user/register/email/activate/resend", handler.UserTokenVerify(), userRouter.ResendActivateEmail)
		apiV1.POST("/user/register/email/activate", userRouter.ActivateEmail)
		apiV1.POST("/user/login/email", userRouter.LoginByEmail)

		apiV1.PUT("/user/base", handler.UserTokenVerify(), userRouter.UpdateUserBaseInfo)
		apiV1.POST("/user/search", handler.UserTokenVerify(), userRouter.UserSearch)

		apiV1.POST("/user/email/bind", handler.UserTokenVerify(), userRouter.SubmitBindEmail)
		apiV1.GET("/user/email/bind/confirm", userRouter.ConfirmBindEmail)
	}

	// register habit related api
	habitRouter := handler.NewHabitRouter()
	{
		apiV1.POST("/habit", handler.UserTokenVerify(), habitRouter.CreateHabit)
		apiV1.GET("habit/:id", handler.UserTokenVerify(), habitRouter.GetHabit)
		apiV1.GET("/habit/list", handler.UserTokenVerify(), habitRouter.ListHabits)
		apiV1.PUT("/habit/:id", handler.UserTokenVerify(), habitRouter.UpdateHabit)
		apiV1.PUT("/habit/custom/:id", handler.UserTokenVerify(), habitRouter.UpdateHabitUserCustomConfig)
		apiV1.POST("/habit/log/:id", handler.UserTokenVerify(), habitRouter.LogHabit)
	}
}
