package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

const (
	SuccessCode int32 = iota
	InvalidParamCode
	InternalServer
	UnauthorizedCode
)

type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func InternalServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ParseError(err))
}

func InvalidParamError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    InvalidParamCode,
		Message: "invalid params, " + message,
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    UnauthorizedCode,
		Message: "unauthorized",
	})
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "success",
		Data:    data,
	})
}

func ParseError(err error) *Response {
	code := InternalServer
	msg := "internal server error"

	if grpcErr, ok := status.FromError(err); ok {
		code = int32(grpcErr.Code())
		msg = grpcErr.Message()
	}

	return &Response{
		Code:    code,
		Message: msg,
	}
}
