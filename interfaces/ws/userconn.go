package ws

import (
	"sync"
	"time"
)

type UserConn interface {
	GetAll(userID string) ([]*Client, bool)
	Get(userID string, platformID int32) (*Client, error)
	Set(userID string, c *Client) error
	DeleteClients(userID string, clients []*Client) (isDeleteUser bool)
	UserState() <-chan UserState
	GetAllUserStatus(deadline time.Time, nowtime time.Time) []UserState
	RecvSubChange(userID string, platformIDs []int32) bool
}

type UserState struct {
	UserID  string
	Online  []int32
	Offline []int32
}

type UserPlatform struct {
	LastUpTime time.Time
	Clients    []*Client
}

func (u *UserPlatform) PlatformIDs() []int32 {
	if len(u.Clients) == 0 {
		return nil
	}
	platformIDs := make([]int32, 0, len(u.Clients))
	for _, client := range u.Clients {
		platformIDs = append(platformIDs, client.PlatformID)
	}
	return platformIDs
}

func (u *UserPlatform) PlatformIDSet() map[int32]struct{} {
	if len(u.Clients) == 0 {
		return nil
	}
	platformIDs := make(map[int32]struct{})
	for _, client := range u.Clients {
		platformIDs[client.PlatformID] = struct{}{}
	}
	return platformIDs
}

type userConn struct {
	lock sync.RWMutex
	data map[string]*UserPlatform
	ch   chan UserState
}

func newUserConn() UserConn {
	return &userConn{
		data: make(map[string]*UserPlatform),
		ch:   make(chan UserState, 10000),
	}
}

func (u *userConn) GetAll(userID string) ([]*Client, bool) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if platform, ok := u.data[userID]; ok {
		return platform.Clients, true
	}

	return nil, false
}

func (u *userConn) Get(userID string, platformID int32) (*Client, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userConn) Set(userID string, c *Client) error {
	//TODO implement me
	panic("implement me")
}

func (u *userConn) DeleteClients(userID string, clients []*Client) (isDeleteUser bool) {
	//TODO implement me
	panic("implement me")
}

func (u *userConn) UserState() <-chan UserState {
	//TODO implement me
	panic("implement me")
}

func (u *userConn) GetAllUserStatus(deadline time.Time, nowtime time.Time) []UserState {
	//TODO implement me
	panic("implement me")
}

func (u *userConn) RecvSubChange(userID string, platformIDs []int32) bool {
	//TODO implement me
	panic("implement me")
}
