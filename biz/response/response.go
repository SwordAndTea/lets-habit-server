package response

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/swordandtea/lets-habit-server/util"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

type ErrorCode int

const (
	ErrorCode_Success              ErrorCode = 0
	ErrorCode_InvalidParam         ErrorCode = 1001
	ErrorCode_UserAuthFail         ErrorCode = 1101
	ErrorCode_UserNoPermission     ErrorCode = 1102
	ErrroCode_InternalUnknownError ErrorCode = 9999
)

// SError service error
type SError interface {
	error
	ErrorCode() ErrorCode
	CodeLine() string
	Message() string
	Cause() error
	Relocation() SError
}

// sErrorImpl the default implement of SError
type sErrorImpl struct {
	errorCode ErrorCode
	codeLine  string
	message   string
	source    error
}

func (i *sErrorImpl) Error() string {
	return fmt.Sprintf("codeline=%s, msg=%s, source=%v", i.codeLine, i.message, i.source)
}

func (i *sErrorImpl) ErrorCode() ErrorCode {
	return i.errorCode
}

func (i *sErrorImpl) CodeLine() string {
	return i.codeLine
}

func (i *sErrorImpl) Message() string {
	return i.message
}

func (i *sErrorImpl) Cause() error {
	return i.source
}

var gopath = path.Join(os.Getenv("GOPATH"), "src") + "/"

func getCodeLine(stackDepth int) string {
	_, file, line, ok := runtime.Caller(stackDepth)
	if !ok {
		file = "???"
		line = 0
	}
	if strings.HasPrefix(file, gopath) {
		file = file[len(gopath):]
	}
	wd, _ := os.Getwd()
	if strings.HasPrefix(file, wd) {
		file = "." + file[len(wd):]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func (i *sErrorImpl) Relocation() SError {
	return &sErrorImpl{
		source:    i.source,
		codeLine:  getCodeLine(2),
		message:   i.message,
		errorCode: i.errorCode,
	}
}

type HTTPResponseMeta struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	URI     string `json:"uri"`
	Elapsed uint64 `json:"elapsed"`
}

// HTTPResponseBody the response body data
type HTTPResponseBody struct {
	Meta *HTTPResponseMeta `json:"meta"`
	Data interface{}       `json:"data"`
}

// HTTPResponse the model to represent a http response
type HTTPResponse struct {
	statusCode int
	body       *HTTPResponseBody
	timer      *util.Timer
	err        error
}

// NewHTTPResponse create a HTTPResponse with default values
func NewHTTPResponse(rc *app.RequestContext) *HTTPResponse {
	return &HTTPResponse{
		statusCode: http.StatusOK,
		body: &HTTPResponseBody{
			Meta: &HTTPResponseMeta{
				Status:  0,
				Message: "success",
				URI:     fmt.Sprintf("%s %s", string(rc.Request.Method()), string(rc.Request.URI().Path())),
				Elapsed: 0,
			},
			Data: nil,
		},
		timer: util.NewTimer(),
		err:   nil,
	}
}

// ReturnWithLog return body with response meta logged
func (r *HTTPResponse) ReturnWithLog(ctx context.Context, rc *app.RequestContext) {
	r.body.Meta.Elapsed = r.timer.ElapsedUInt64()
	rc.JSON(r.statusCode, r.body)
	if r.err != nil {
		hlog.Errorf("%s %d, elapsed: %d, err=%v", r.body.Meta.URI, r.statusCode, r.body.Meta.Elapsed, r.err)
	} else {
		hlog.Infof("%s %d, elapsed: %d", r.body.Meta.URI, r.statusCode, r.body.Meta.Elapsed)
	}
}

// Abort stop the call chain and return with http status and http body
func (r *HTTPResponse) Abort(ctx context.Context, rc *app.RequestContext) {
	r.body.Meta.Elapsed = r.timer.ElapsedUInt64()
	rc.AbortWithStatusJSON(r.statusCode, r.body)
	if r.err != nil {
		hlog.Errorf("%s %d, elapsed: %d, err=%v", r.body.Meta.URI, r.statusCode, r.body.Meta.Elapsed, r.err)
	} else {
		hlog.Infof("%s %d, elapsed: %d", r.body.Meta.URI, r.statusCode, r.body.Meta.Elapsed)
	}
}

// SetSuccessData mark request success by fill meta field with success message and set body data
func (r *HTTPResponse) SetSuccessData(data interface{}) {
	r.statusCode = http.StatusOK
	r.body.Meta.Status = int(ErrorCode_Success)
	r.body.Meta.Message = "success"
	r.body.Data = data
	r.err = nil
}

// SetError set request fail by fill meta filed with error message,
// if err is type of SError, do some special action like auto set http status by ErrorCode value
func (r *HTTPResponse) SetError(err error) {
	switch err.(type) {
	case SError:
		sErr := err.(SError)
		switch sErr.ErrorCode() {
		case ErrorCode_InvalidParam:
			r.statusCode = http.StatusBadRequest
		case ErrorCode_UserAuthFail, ErrorCode_UserNoPermission:
			r.statusCode = http.StatusForbidden
		default:
			r.statusCode = http.StatusInternalServerError
		}
		r.body.Meta.Status = int(sErr.ErrorCode())
		r.body.Meta.Message = sErr.Message()
	default:
		r.statusCode = http.StatusInternalServerError
		r.body.Meta.Status = int(ErrroCode_InternalUnknownError)
		r.body.Meta.Message = "internal unknown error"
	}
	r.err = err
}

// New create a new SError with specific ErrorCode without new origin error
func (ec ErrorCode) New(format string, args ...interface{}) SError {
	return &sErrorImpl{
		errorCode: ec,
		codeLine:  getCodeLine(2),
		message:   fmt.Sprintf(format, args...),
		source:    nil,
	}
}

// Wrap create a new SError by wrap an origin error with specific ErrorCode
func (ec ErrorCode) Wrap(err error, extraMsg string, args ...interface{}) SError {
	return &sErrorImpl{
		errorCode: ec,
		codeLine:  getCodeLine(2),
		message:   fmt.Sprintf(extraMsg, args...),
		source:    err,
	}
}
