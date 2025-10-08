package ws

import (
	"context"
	"sync/atomic"

	wsctx "github.com/crazyfrankie/goim/interfaces/ws/context"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func (s *Server) InitServer(ctx context.Context) error {
	// 初始化服务器
	logs.Infof("InitServer completed")
	return nil
}

type Server struct {
	LongConnServer LongConnServer
	pushTerminal   map[int]struct{}
	ready          func(srv *Server) error
}

func (s *Server) SetLongConnServer(LongConnServer LongConnServer) {
	s.LongConnServer = LongConnServer
}

func NewServer(longConnServer LongConnServer, ready func(srv *Server) error) *Server {
	s := &Server{
		LongConnServer: longConnServer,
		pushTerminal:   make(map[int]struct{}),
		ready:          ready,
	}
	// 设置推送终端
	s.pushTerminal[2] = struct{}{} // iOS
	s.pushTerminal[3] = struct{}{} // Android
	return s
}

func (s *Server) GetUsersOnlineStatus(ctx context.Context, req *GetUsersOnlineStatusReq) (*GetUsersOnlineStatusResp, error) {
	var resp GetUsersOnlineStatusResp
	for _, userID := range req.UserIDs {
		clients, ok := s.LongConnServer.GetUserAllCons(userID)
		if !ok {
			continue
		}

		uresp := new(GetUsersOnlineStatusResp_SuccessResult)
		uresp.UserID = userID
		for _, client := range clients {
			if client == nil {
				continue
			}

			ps := new(GetUsersOnlineStatusResp_SuccessDetail)
			ps.PlatformID = int32(client.PlatformID)
			ps.ConnID = client.ctx.GetConnID()
			ps.Token = client.Token
			ps.IsBackground = client.IsBackground
			uresp.Status = 1 // Online
			uresp.DetailPlatformStatus = append(uresp.DetailPlatformStatus, ps)
		}
		if uresp.Status == 1 {
			resp.SuccessResult = append(resp.SuccessResult, uresp)
		}
	}
	return &resp, nil
}

func (s *Server) pushToUser(ctx context.Context, userID string, msgData []byte) *SingleMsgToUserResults {
	clients, ok := s.LongConnServer.GetUserAllCons(userID)
	if !ok {
		logs.Debugf("push user not online, userID: %s", userID)
		return &SingleMsgToUserResults{
			UserID: userID,
		}
	}

	logs.Debugf("push user online, clients count: %d, userID: %s", len(clients), userID)
	result := &SingleMsgToUserResults{
		UserID: userID,
		Resp:   make([]*SingleMsgToUserPlatform, 0, len(clients)),
	}

	for _, client := range clients {
		if client == nil {
			continue
		}
		userPlatform := &SingleMsgToUserPlatform{
			RecvPlatFormID: int32(client.PlatformID),
		}

		if !client.IsBackground || (client.IsBackground && client.PlatformID != 2) { // iOS平台ID为2
			err := client.PushMessage(ctx, msgData)
			if err != nil {
				logs.Warnf("online push msg failed, userID: %s, platformID: %d, err: %v", userID, client.PlatformID, err)
				userPlatform.ResultCode = int64(500) // 错误码
			} else {
				if _, ok := s.pushTerminal[int(client.PlatformID)]; ok {
					result.OnlinePush = true
				}
			}
		} else {
			userPlatform.ResultCode = int64(501) // iOS后台推送错误码
		}
		result.Resp = append(result.Resp, userPlatform)
	}
	return result
}

func (s *Server) SuperGroupOnlineBatchPushOneMsg(ctx context.Context, req *OnlineBatchPushOneMsgReq) (*OnlineBatchPushOneMsgResp, error) {
	if len(req.PushToUserIDs) == 0 {
		return &OnlineBatchPushOneMsgResp{}, nil
	}

	ch := make(chan *SingleMsgToUserResults, len(req.PushToUserIDs))
	var count atomic.Int64
	count.Add(int64(len(req.PushToUserIDs)))

	for i := range req.PushToUserIDs {
		userID := req.PushToUserIDs[i]
		go func(uid string) {
			ch <- s.pushToUser(ctx, uid, req.MsgData)
			if count.Add(-1) == 0 {
				close(ch)
			}
		}(userID)
	}

	resp := &OnlineBatchPushOneMsgResp{
		SinglePushResult: make([]*SingleMsgToUserResults, 0, len(req.PushToUserIDs)),
	}

	for {
		select {
		case <-ctx.Done():
			logs.Errorf("SuperGroupOnlineBatchPushOneMsg ctx done: %v", context.Cause(ctx))
			return resp, nil
		case res, ok := <-ch:
			if !ok {
				return resp, nil
			}
			resp.SinglePushResult = append(resp.SinglePushResult, res)
		}
	}
}

func (s *Server) KickUserOffline(ctx context.Context, req *KickUserOfflineReq) (*KickUserOfflineResp, error) {
	for _, v := range req.KickUserIDList {
		clients, _, ok := s.LongConnServer.GetUserPlatformCons(v, int(req.PlatformID))
		if !ok {
			logs.Debugf("conn not exist, userID: %s, platformID: %d", v, req.PlatformID)
			continue
		}

		for _, client := range clients {
			logs.Debugf("kick user offline, userID: %s, platformID: %d", v, req.PlatformID)
			if err := client.ConnServer.KickUserConn(client); err != nil {
				logs.Warnf("kick user offline failed, userID: %s, platformID: %d, err: %v", v, req.PlatformID, err)
			}
		}
	}

	return &KickUserOfflineResp{}, nil
}

func (s *Server) MultiTerminalLoginCheck(ctx context.Context, req *MultiTerminalLoginCheckReq) (*MultiTerminalLoginCheckResp, error) {
	if oldClients, userOK, clientOK := s.LongConnServer.GetUserPlatformCons(req.UserID, int(req.PlatformID)); userOK {
		tempUserCtx := wsctx.NewTempContext()
		tempUserCtx.SetToken(req.Token)
		client := &Client{}
		client.ctx = tempUserCtx
		client.Token = req.Token
		client.UserID = req.UserID
		client.PlatformID = req.PlatformID
		i := &kickHandler{
			clientOK:   clientOK,
			oldClients: oldClients,
			newClient:  client,
		}
		s.LongConnServer.SetKickHandlerInfo(i)
	}
	return &MultiTerminalLoginCheckResp{}, nil
}

// 相关数据结构
type GetUsersOnlineStatusReq struct {
	UserIDs []string `json:"userIDs"`
}

type GetUsersOnlineStatusResp struct {
	SuccessResult []*GetUsersOnlineStatusResp_SuccessResult `json:"successResult"`
}

type GetUsersOnlineStatusResp_SuccessResult struct {
	UserID               string                                    `json:"userID"`
	Status               int32                                     `json:"status"`
	DetailPlatformStatus []*GetUsersOnlineStatusResp_SuccessDetail `json:"detailPlatformStatus"`
}

type GetUsersOnlineStatusResp_SuccessDetail struct {
	PlatformID   int32  `json:"platformID"`
	ConnID       string `json:"connID"`
	Token        string `json:"token"`
	IsBackground bool   `json:"isBackground"`
}

type OnlineBatchPushOneMsgReq struct {
	PushToUserIDs []string `json:"pushToUserIDs"`
	MsgData       []byte   `json:"msgData"`
}

type OnlineBatchPushOneMsgResp struct {
	SinglePushResult []*SingleMsgToUserResults `json:"singlePushResult"`
}

type SingleMsgToUserResults struct {
	UserID     string                     `json:"userID"`
	Resp       []*SingleMsgToUserPlatform `json:"resp"`
	OnlinePush bool                       `json:"onlinePush"`
}

type SingleMsgToUserPlatform struct {
	RecvPlatFormID int32 `json:"recvPlatFormID"`
	ResultCode     int64 `json:"resultCode"`
}

type KickUserOfflineReq struct {
	KickUserIDList []string `json:"kickUserIDList"`
	PlatformID     int32    `json:"platformID"`
}

type KickUserOfflineResp struct{}

type MultiTerminalLoginCheckReq struct {
	UserID     string `json:"userID"`
	PlatformID int32  `json:"platformID"`
	Token      string `json:"token"`
}

type MultiTerminalLoginCheckResp struct{}
