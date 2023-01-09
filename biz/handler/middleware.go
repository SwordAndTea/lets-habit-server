package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v4"
	"github.com/swordandtea/lets-habit-server/biz/config"
	"github.com/swordandtea/lets-habit-server/biz/response"
)

const UIDKey = "uid"
const UserTokenHeader = "x-lh-auth"

// UserTokenVerify verify user token, return user auth fail error if verify fail, set uid to RequestContext if success
func UserTokenVerify() app.HandlerFunc {
	return func(ctx context.Context, rc *app.RequestContext) {
		resp := response.NewHTTPResponse(rc)
		userToken := string(rc.GetHeader(UserTokenHeader))
		if userToken == "" {
			resp.SetError(response.ErrorCode_UserAuthFail.New("no user token found"))
			resp.Abort(ctx, rc)
			return
		}

		claims := &jwt.RegisteredClaims{}
		_, err := jwt.ParseWithClaims(userToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GlobalConfig.JWT.Cypher), nil
		})

		if err != nil {
			resp.SetError(response.ErrorCode_UserAuthFail.Wrap(err, "verify user token fail"))
			resp.Abort(ctx, rc)
			return
		}

		if claims.ID == "" {
			resp.SetError(response.ErrorCode_UserAuthFail.Wrap(err, "invalid user token, no user id found"))
			resp.Abort(ctx, rc)
			return
		}

		rc.Set(UIDKey, claims.ID)
	}
}
