package ws

import (
	"context"
	"sync"

	"github.com/crazyfrankie/goim/pkg/sonic"
	"github.com/go-playground/validator/v10"
)

const (
	TextPing = "ping"
	TextPong = "pong"
)

type TextMessage struct {
	Type string `json:"type"`
	Body []byte `json:"body"`
}

type MessageHandler interface {
	GetSeq(ctx context.Context, data *Req) ([]byte, error)
	SendMessage(ctx context.Context, data *Req) ([]byte, error)
	SendSignalMessage(ctx context.Context, data *Req) ([]byte, error)
	PullMessageBySeqList(ctx context.Context, data *Req) ([]byte, error)
	GetConversationsHasReadAndMaxSeq(ctx context.Context, data *Req) ([]byte, error)
	GetSeqMessage(ctx context.Context, data *Req) ([]byte, error)
	UserLogout(ctx context.Context, data *Req) ([]byte, error)
	SetUserDeviceBackground(ctx context.Context, data *Req) ([]byte, bool, error)
	GetLastMessage(ctx context.Context, data *Req) ([]byte, error)
}

type Req struct {
	ReqIdentifier int32  `json:"reqIdentifier"`
	Token         string `json:"token"`
	SendID        string `json:"sendID"`
	OperationID   string `json:"operationID"`
	MsgIncr       string `json:"msgIncr"`
	Data          []byte `json:"data"`
}

func (r *Req) String() string {
	var tReq Req
	tReq.ReqIdentifier = r.ReqIdentifier
	tReq.Token = r.Token
	tReq.SendID = r.SendID
	tReq.OperationID = r.OperationID
	tReq.MsgIncr = r.MsgIncr
	return structToJSONStr(tReq)
}

var reqPool = sync.Pool{
	New: func() any {
		return new(Req)
	},
}

func getReq() *Req {
	req := reqPool.Get().(*Req)
	req.Data = nil
	req.MsgIncr = ""
	req.OperationID = ""
	req.ReqIdentifier = 0
	req.SendID = ""
	req.Token = ""
	return req
}

func freeReq(req *Req) {
	reqPool.Put(req)
}

type Resp struct {
	ReqIdentifier int32       `json:"reqIdentifier"`
	MsgIncr       string      `json:"msgIncr"`
	OperationID   string      `json:"operationID"`
	ErrCode       int32       `json:"errCode"`
	ErrMsg        string      `json:"errMsg"`
	Data          interface{} `json:"data"`
}

func (r *Resp) String() string {
	var tResp Resp
	tResp.ReqIdentifier = r.ReqIdentifier
	tResp.MsgIncr = r.MsgIncr
	tResp.OperationID = r.OperationID
	tResp.ErrCode = r.ErrCode
	tResp.ErrMsg = r.ErrMsg
	return structToJSONStr(tResp)
}

type GrpcHandler struct {
	validate *validator.Validate
	// msgClient    pb.MsgClient
	// pushClient   pb.PushClient
	// userClient   pb.UserClient
}

func NewGrpcHandler(validate *validator.Validate) *GrpcHandler {
	return &GrpcHandler{
		validate: validate,
	}
}

func (g *GrpcHandler) GetSeq(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) SendMessage(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) SendSignalMessage(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) PullMessageBySeqList(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) UserLogout(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) SetUserDeviceBackground(ctx context.Context, data *Req) ([]byte, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) GetConversationsHasReadAndMaxSeq(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) GetSeqMessage(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GrpcHandler) GetLastMessage(ctx context.Context, data *Req) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func structToJSONStr(d any) string {
	res, _ := sonic.MarshalString(d)
	return res
}
