package context

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/crazyfrankie/goim/interfaces/ws/types"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	"github.com/crazyfrankie/goim/pkg/lang/encrypt"
	"github.com/crazyfrankie/goim/types/errno"
)

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	RemoteAddr string
	ConnID     string
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Value(key any) any {
	switch key {
	case types.OpUserID:
		return c.GetUserID()
	case types.OperationID:
		return c.GetOperationID()
	case types.ConnID:
		return c.GetConnID()
	case types.RemoteAddr:
		return c.RemoteAddr
	default:
		return ""
	}
}

func NewContext(r http.ResponseWriter, req *http.Request) *Context {
	remoteAddr := req.RemoteAddr
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr += "_" + forwarded
	}
	return &Context{
		Writer:     r,
		Request:    req,
		Path:       req.URL.Path,
		Method:     req.Method,
		RemoteAddr: remoteAddr,
		ConnID:     userConnID(remoteAddr),
	}
}

func NewTempContext() *Context {
	return &Context{
		Request: &http.Request{URL: &url.URL{}},
	}
}

func (c *Context) GetRemoteAddr() string {
	return c.RemoteAddr
}

func (c *Context) Query(key string) (string, bool) {
	var value string
	if value = c.Request.URL.Query().Get(key); value == "" {
		return value, false
	}
	return value, true
}

func (c *Context) GetHeader(key string) (string, bool) {
	var value string
	if value = c.Request.Header.Get(key); value == "" {
		return value, false
	}
	return value, true
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) GetConnID() string {
	return c.ConnID
}

func (c *Context) GetUserID() string {
	return c.Request.URL.Query().Get(types.WsUserID)
}

func (c *Context) GetPlatformID() string {
	return c.Request.URL.Query().Get(types.PlatformID)
}

func (c *Context) GetOperationID() string {
	return c.Request.URL.Query().Get(types.OperationID)
}

func (c *Context) SetOperationID(operationID string) {
	values := c.Request.URL.Query()
	values.Set(types.OperationID, operationID)
	c.Request.URL.RawQuery = values.Encode()
}

func (c *Context) GetToken() string {
	return c.Request.URL.Query().Get(types.Token)
}

func (c *Context) SetToken(token string) {
	c.Request.URL.RawQuery = types.Token + "=" + token
}

func (c *Context) GetCompression() bool {
	compression, exists := c.Query(types.Compression)
	if exists && compression == types.GzipCompressionProtocol {
		return true
	} else {
		compression, exists := c.GetHeader(types.Compression)
		if exists && compression == types.GzipCompressionProtocol {
			return true
		}
	}
	return false
}

func (c *Context) GetSDKType() string {
	sdkType := c.Request.URL.Query().Get(types.SDKType)
	if sdkType == "" {
		sdkType = types.GoSDK
	}
	return sdkType
}

func (c *Context) ShouldSendResp() bool {
	errResp, exists := c.Query(types.SendResponse)
	if exists {
		b, err := strconv.ParseBool(errResp)
		if err != nil {
			return false
		} else {
			return b
		}
	}
	return false
}

func (c *Context) ParseEssentialArgs() error {
	_, exists := c.Query(types.Token)
	if !exists {
		return errorx.New(errno.ErrConnArgsCode, errorx.KV("cause", "token is empty"))
	}
	_, exists = c.Query(types.WsUserID)
	if !exists {
		return errorx.New(errno.ErrConnArgsCode, errorx.KV("cause", "sendID is empty"))
	}
	platformIDStr, exists := c.Query(types.PlatformID)
	if !exists {
		return errorx.New(errno.ErrConnArgsCode, errorx.KV("cause", "platformID is empty"))
	}
	_, err := strconv.Atoi(platformIDStr)
	if err != nil {
		return errorx.New(errno.ErrConnArgsCode, errorx.KV("cause", "platformID is not int"))
	}
	switch sdkType, _ := c.Query(types.SDKType); sdkType {
	case "", types.GoSDK, types.JsSDK:
	default:
		return errorx.New(errno.ErrConnArgsCode, errorx.KV("cause", "sdkType is not go or js"))
	}
	return nil
}

func userConnID(remoteAddr string) string {
	return encrypt.Md5(remoteAddr + "_" + conv.Int64ToStr(time.Now().UnixMilli()))
}
