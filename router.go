// Code generated by hertz generator.

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/swordandtea/fhwh/biz/handler"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)

	// your code ...
	apiV1 := r.Group("api/v1")

	// register user related api
	userRouter := handler.NewUserRouter()
	{
		apiV1.POST("/user/register/email", userRouter.RegisterByEmail)
		apiV1.GET("/user/register/email/activate", userRouter.ActivateEmail)

		apiV1.POST("/user/email/bind", handler.UserTokenVerify(), userRouter.SubmitBindEmail)
		apiV1.GET("/user/email/bind/confirm", userRouter.ConfirmBindEmail)
	}
}
