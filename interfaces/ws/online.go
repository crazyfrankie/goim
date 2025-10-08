package ws

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/crazyfrankie/goim/pkg/logs"
)

func (ws *WebsocketServer) ChangeOnlineStatus(concurrent int) {
	if concurrent < 1 {
		concurrent = 1
	}
	const renewalTime = time.Minute * 5 // 5分钟续期时间
	renewalTicker := time.NewTicker(renewalTime)

	requestChs := make([]chan *SetUserOnlineStatusReq, concurrent)
	changeStatus := make([][]UserState, concurrent)

	for i := 0; i < concurrent; i++ {
		requestChs[i] = make(chan *SetUserOnlineStatusReq, 64)
		changeStatus[i] = make([]UserState, 0, 100)
	}

	mergeTicker := time.NewTicker(time.Second)

	local2pb := func(u UserState) *UserOnlineStatus {
		return &UserOnlineStatus{
			UserID:  u.UserID,
			Online:  u.Online,
			Offline: u.Offline,
		}
	}

	rNum := rand.Uint64()
	pushUserState := func(us ...UserState) {
		for _, u := range us {
			sum := md5.Sum([]byte(u.UserID))
			i := (binary.BigEndian.Uint64(sum[:]) + rNum) % uint64(concurrent)
			changeStatus[i] = append(changeStatus[i], u)
			status := changeStatus[i]
			if len(status) == cap(status) {
				req := &SetUserOnlineStatusReq{
					Status: make([]*UserOnlineStatus, len(status)),
				}
				for j, s := range status {
					req.Status[j] = local2pb(s)
				}
				changeStatus[i] = status[:0]
				select {
				case requestChs[i] <- req:
				default:
					logs.Errorf("user online processing is too slow")
				}
			}
		}
	}

	pushAllUserState := func() {
		for i, status := range changeStatus {
			if len(status) == 0 {
				continue
			}
			req := &SetUserOnlineStatusReq{
				Status: make([]*UserOnlineStatus, len(status)),
			}
			for j, s := range status {
				req.Status[j] = local2pb(s)
			}
			changeStatus[i] = status[:0]
			select {
			case requestChs[i] <- req:
			default:
				logs.Errorf("user online processing is too slow")
			}
		}
	}

	var count atomic.Int64
	operationIDPrefix := fmt.Sprintf("p_%d_", os.Getpid())
	doRequest := func(req *SetUserOnlineStatusReq) {
		opID := operationIDPrefix + strconv.FormatInt(count.Add(1), 10)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// 这里应该调用用户服务更新在线状态
		// 由于我们没有用户服务客户端，这里简化处理
		logs.Debugf("update user online status, operationID: %s, count: %d", opID, len(req.Status))

		for _, ss := range req.Status {
			for _, online := range ss.Online {
				clients, _, _ := ws.GetUserPlatformCons(ss.UserID, int(online))
				back := false
				if len(clients) > 0 {
					back = clients[0].IsBackground
				}
				// 这里应该调用webhook
				logs.Debugf("user online webhook, userID: %s, platformID: %d, background: %v", ss.UserID, online, back)
			}
			for _, offline := range ss.Offline {
				// 这里应该调用webhook
				logs.Debugf("user offline webhook, userID: %s, platformID: %d", ss.UserID, offline)
			}
		}
	}

	for i := 0; i < concurrent; i++ {
		go func(ch <-chan *SetUserOnlineStatusReq) {
			for req := range ch {
				doRequest(req)
			}
		}(requestChs[i])
	}

	for {
		select {
		case <-mergeTicker.C:
			pushAllUserState()
		case now := <-renewalTicker.C:
			deadline := now.Add(-time.Minute * 5)
			users := ws.getAllUserStatus(deadline, now)
			logs.Debugf("renewal ticker, deadline: %v, nowtime: %v, num: %d", deadline, now, len(users))
			pushUserState(users...)
		case state := <-ws.getUserStateChannel():
			logs.Debugf("OnlineCache user online change, userID: %s, online: %v, offline: %v", state.UserID, state.Online, state.Offline)
			pushUserState(state)
		}
	}
}

// getAllUserStatus 获取所有用户状态
func (ws *WebsocketServer) getAllUserStatus(deadline time.Time, nowtime time.Time) []UserState {
	var result []UserState
	buckets := ws.bucketManager.GetAllBuckets()

	userPlatforms := make(map[string]map[int32]struct{})

	for _, bucket := range buckets {
		bucket.lock.RLock()
		for userID, userPlatform := range bucket.userMap {
			userPlatform.mutex.RLock()
			if deadline.Before(time.Unix(userPlatform.lastTime, 0)) {
				userPlatform.mutex.RUnlock()
				continue
			}

			if userPlatforms[userID] == nil {
				userPlatforms[userID] = make(map[int32]struct{})
			}

			for platformID := range userPlatform.Platforms {
				userPlatforms[userID][platformID] = struct{}{}
			}
			userPlatform.mutex.RUnlock()
		}
		bucket.lock.RUnlock()
	}

	for userID, platforms := range userPlatforms {
		online := make([]int32, 0, len(platforms))
		for platformID := range platforms {
			online = append(online, platformID)
		}
		result = append(result, UserState{UserID: userID, Online: online})
	}

	return result
}

// getUserStateChannel 获取用户状态变化通道
func (ws *WebsocketServer) getUserStateChannel() <-chan UserState {
	// 这里应该返回一个用户状态变化的通道
	// 由于我们使用BucketManager，需要实现一个状态变化通知机制
	// 这里简化实现，返回一个空通道
	ch := make(chan UserState)
	close(ch)
	return ch
}

// 相关数据结构
type SetUserOnlineStatusReq struct {
	Status []*UserOnlineStatus `json:"status"`
}

type UserOnlineStatus struct {
	UserID  string  `json:"userID"`
	Online  []int32 `json:"online"`
	Offline []int32 `json:"offline"`
	ConnID  string  `json:"connID,omitempty"`
}

type UserState struct {
	UserID  string
	Online  []int32
	Offline []int32
}
