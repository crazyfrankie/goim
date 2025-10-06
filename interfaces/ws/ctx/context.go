package ctx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/crazyfrankie/goim/interfaces/ws/constant"
	"github.com/crazyfrankie/goim/pkg/lang/encrypt"
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
	case constant.OpUserID:
		return c.GetUserID()
	case constant.OperationID:
		return c.GetOperationID()
	case constant.ConnID:
		return c.GetConnID()
	case constant.RemoteAddr:
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
	return c.Request.URL.Query().Get(constant.WsUserID)
}

func (c *Context) GetPlatformID() string {
	return c.Request.URL.Query().Get(constant.PlatformID)
}

func (c *Context) GetOperationID() string {
	return c.Request.URL.Query().Get(constant.OperationID)
}

func (c *Context) SetOperationID(operationID string) {
	values := c.Request.URL.Query()
	values.Set(constant.OperationID, operationID)
	c.Request.URL.RawQuery = values.Encode()
}

func (c *Context) GetToken() string {
	return c.Request.URL.Query().Get(constant.Token)
}

func (c *Context) GetCompression() bool {
	compression, exists := c.Query(constant.Compression)
	if exists && compression == constant.GzipCompressionProtocol {
		return true
	} else {
		compression, exists := c.GetHeader(constant.Compression)
		if exists && compression == constant.GzipCompressionProtocol {
			return true
		}
	}
	return false
}

func (c *Context) GetSDKType() string {
	sdkType := c.Request.URL.Query().Get(constant.SDKType)
	if sdkType == "" {
		sdkType = constant.GoSDK
	}
	return sdkType
}

func (c *Context) ShouldSendResp() bool {
	errResp, exists := c.Query(constant.SendResponse)
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

func userConnID(remoteAddr string) string {
	return encrypt.Md5(remoteAddr + "_" + strconv.FormatInt(time.Now().UnixMilli(), 10))
}
