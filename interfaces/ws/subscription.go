package ws

import (
	"context"
	"sync"

	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/sonic"
)

func (ws *WebsocketServer) subscriberUserOnlineStatusChanges(ctx context.Context, userID string, platformIDs []int32) {
	// 检查是否有用户订阅了这个用户的状态变化
	// 这里简化实现，实际应该检查bucketManager中的连接
	logs.Debugf("gateway receive subscription message and go back online, userID: %s, platformIDs: %v", userID, platformIDs)
	ws.pushUserIDOnlineStatus(ctx, userID, platformIDs)
}

func (ws *WebsocketServer) SubUserOnlineStatus(ctx context.Context, client *Client, data *Req) ([]byte, error) {
	var sub SubUserOnlineStatus
	if err := sonic.Unmarshal(data.Data, &sub); err != nil {
		return nil, err
	}

	ws.subscription.Sub(client, sub.SubscribeUserID, sub.UnsubscribeUserID)

	var resp SubUserOnlineStatusTips
	if len(sub.SubscribeUserID) > 0 {
		resp.Subscribers = make([]*SubUserOnlineStatusElem, 0, len(sub.SubscribeUserID))
		for _, userID := range sub.SubscribeUserID {
			// 获取用户在线平台信息
			platformIDs := ws.getUserOnlinePlatforms(userID)
			resp.Subscribers = append(resp.Subscribers, &SubUserOnlineStatusElem{
				UserID:            userID,
				OnlinePlatformIDs: platformIDs,
			})
		}
	}

	return sonic.Marshal(&resp)
}

// getUserOnlinePlatforms 获取用户在线平台列表
func (ws *WebsocketServer) getUserOnlinePlatforms(userID string) []int32 {
	var platformIDs []int32
	buckets := ws.bucketManager.GetAllBuckets()

	platformSet := make(map[int32]struct{})
	for _, bucket := range buckets {
		clients := bucket.GetUserClients(userID)
		for _, client := range clients {
			platformSet[client.PlatformID] = struct{}{}
		}
	}

	for platformID := range platformSet {
		platformIDs = append(platformIDs, platformID)
	}

	return platformIDs
}

func newSubscription() *Subscription {
	return &Subscription{
		userIDs: make(map[string]*subClient),
	}
}

type subClient struct {
	clients map[string]*Client
}

type Subscription struct {
	lock    sync.RWMutex
	userIDs map[string]*subClient // subscribe to the user's client connection
}

func (s *Subscription) DelClient(client *Client) {
	client.subLock.Lock()
	userIDs := make([]string, 0, len(client.subscriptions))
	for userID := range client.subscriptions {
		userIDs = append(userIDs, userID)
		delete(client.subscriptions, userID)
	}
	client.subLock.Unlock()

	if len(userIDs) == 0 {
		return
	}

	addr := client.ctx.GetRemoteAddr()
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, userID := range userIDs {
		sub, ok := s.userIDs[userID]
		if !ok {
			continue
		}
		delete(sub.clients, addr)
		if len(sub.clients) == 0 {
			delete(s.userIDs, userID)
		}
	}
}

func (s *Subscription) GetClient(userID string) []*Client {
	s.lock.RLock()
	defer s.lock.RUnlock()

	cs, ok := s.userIDs[userID]
	if !ok {
		return nil
	}

	clients := make([]*Client, 0, len(cs.clients))
	for _, client := range cs.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *Subscription) Sub(client *Client, addUserIDs, delUserIDs []string) {
	if len(addUserIDs)+len(delUserIDs) == 0 {
		return
	}

	var (
		del = make(map[string]struct{})
		add = make(map[string]struct{})
	)

	client.subLock.Lock()
	for _, userID := range delUserIDs {
		if _, ok := client.subscriptions[userID]; !ok {
			continue
		}
		del[userID] = struct{}{}
		delete(client.subscriptions, userID)
	}

	for _, userID := range addUserIDs {
		delete(del, userID)
		if _, ok := client.subscriptions[userID]; ok {
			continue
		}
		client.subscriptions[userID] = struct{}{}
		add[userID] = struct{}{}
	}
	client.subLock.Unlock()

	if len(del)+len(add) == 0 {
		return
	}

	addr := client.ctx.GetRemoteAddr()
	s.lock.Lock()
	defer s.lock.Unlock()

	for userID := range del {
		sub, ok := s.userIDs[userID]
		if !ok {
			continue
		}
		delete(sub.clients, addr)
		if len(sub.clients) == 0 {
			delete(s.userIDs, userID)
		}
	}

	for userID := range add {
		sub, ok := s.userIDs[userID]
		if !ok {
			sub = &subClient{clients: make(map[string]*Client)}
			s.userIDs[userID] = sub
		}
		sub.clients[addr] = client
	}
}

func (ws *WebsocketServer) pushUserIDOnlineStatus(ctx context.Context, userID string, platformIDs []int32) {
	clients := ws.subscription.GetClient(userID)
	if len(clients) == 0 {
		return
	}

	onlineStatus := &SubUserOnlineStatusTips{
		Subscribers: []*SubUserOnlineStatusElem{{UserID: userID, OnlinePlatformIDs: platformIDs}},
	}

	statusData, err := sonic.Marshal(onlineStatus)
	if err != nil {
		logs.Errorf("pushUserIDOnlineStatus json.Marshal failed: %v", err)
		return
	}

	for _, client := range clients {
		if err := client.PushUserOnlineStatus(statusData); err != nil {
			logs.Errorf("UserSubscribeOnlineStatusNotification push failed: %v, userID: %s, platformID: %d, changeUserID: %s, changePlatformID: %v",
				err, client.UserID, client.PlatformID, userID, platformIDs)
		}
	}
}

// 订阅相关的数据结构
type SubUserOnlineStatus struct {
	SubscribeUserID   []string `json:"subscribeUserID"`
	UnsubscribeUserID []string `json:"unsubscribeUserID"`
}

type SubUserOnlineStatusTips struct {
	Subscribers []*SubUserOnlineStatusElem `json:"subscribers"`
}

type SubUserOnlineStatusElem struct {
	UserID            string  `json:"userID"`
	OnlinePlatformIDs []int32 `json:"onlinePlatformIDs"`
}
