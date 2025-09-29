package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func InternalServerError(c *gin.Context, err error) {
	if grpcErr, ok := status.FromError(err); ok {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    int32(grpcErr.Code()),
			Message: grpcErr.Message(),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: "internal server error",
	})
}

func InvalidParamError(c *gin.Context) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Message: "invalid params",
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: "unauthorized",
	})
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
