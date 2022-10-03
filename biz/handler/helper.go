package handler

import (
	"github.com/swordandtea/fhwh/biz/response"
	"regexp"
)

func BindAndValidateErr(err error) response.SError {
	return response.ErrorCode_InvalidParam.Wrap(err, "bind req fail")
}

func ValidateEmail(e string) response.SError {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(e) {
		return response.ErrorCode_InvalidParam.New("invalid email")
	}
	return nil
}

func ValidatePassword(p string) response.SError {
	if len(p) < 8 {
		return response.ErrorCode_InvalidParam.New("password length is less than eight")
	}
	return nil
}
