package ws

import (
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"
)

// BroadcastReq represents a broadcast request
type BroadcastReq struct {
	RoomID  string
	Message []byte
}

// BucketManager performs sharding based on city-hash, where the number of cities (bucketNum) can be specified.
// Additionally, we implement secondary sharding to account for users potentially logging in across different platforms,
// we manage each user's connections across multiple platforms using UserID as the key and UserPlatforms as the value.
type BucketManager struct {
	buckets   []*Bucket
	bucketNum uint32
}

type BucketConfig struct {
	ChannelSize   int
	RoomSize      int
	RoutineAmount int
	RoutineSize   int
}

func DefaultBucketConfig() *BucketConfig {
	return &BucketConfig{
		ChannelSize:   0,
		RoomSize:      0,
		RoutineAmount: 0,
		RoutineSize:   0,
	}
}

type ServerConfig struct {
	BucketNum     int
	MaxConnNum    int64
	HeartbeatTime time.Duration
	Bucket        *BucketConfig
	Client        *ClientConfig
}

func NewBucketManager(bucketNum int, config *BucketConfig) *BucketManager {
	bm := &BucketManager{
		buckets:   make([]*Bucket, bucketNum),
		bucketNum: uint32(bucketNum),
	}

	for i := 0; i < bucketNum; i++ {
		bm.buckets[i] = NewBucket(i, config)
	}

	return bm
}

func (bm *BucketManager) GetBucket(key string) *Bucket {
	hash := crc32.ChecksumIEEE([]byte(key))
	return bm.buckets[hash%bm.bucketNum]
}

func (bm *BucketManager) GetAllBuckets() []*Bucket {
	return bm.buckets
}

// Bucket Connecting Shards
type Bucket struct {
	id   int
	lock sync.RWMutex

	// Connection Management
	clients map[string]*Client
	rooms   map[string]*Room

	// User Management
	userMap map[string]*UserPlatforms
	ipCount map[string]int32

	// Asynchronous processing
	routines   []chan *BroadcastReq
	routineNum uint64
}

type UserPlatforms struct {
	UserID    string
	Platforms map[int32][]*Client
	mutex     sync.RWMutex
	lastTime  int64
}

func NewBucket(id int, config *BucketConfig) *Bucket {
	b := &Bucket{
		id:       id,
		clients:  make(map[string]*Client, config.ChannelSize),
		rooms:    make(map[string]*Room, config.RoomSize),
		userMap:  make(map[string]*UserPlatforms),
		ipCount:  make(map[string]int32),
		routines: make([]chan *BroadcastReq, config.RoutineAmount),
	}

	for i := 0; i < config.RoutineAmount; i++ {
		ch := make(chan *BroadcastReq, config.RoutineSize)
		b.routines[i] = ch
		go b.roomProcessor(ch)
	}

	return b
}

// PutClient Add client connection
func (b *Bucket) PutClient(client *Client) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	// 关闭旧连接
	if oldClient := b.clients[client.Key()]; oldClient != nil {
		oldClient.close()
	}

	// 添加新连接
	b.clients[client.Key()] = client

	// 更新用户平台映射
	b.updateUserPlatforms(client, true)

	// 更新IP统计
	b.ipCount[client.IP()]++

	return nil
}

// DelClient Delete Client connection
func (b *Bucket) DelClient(client *Client) {
	b.lock.Lock()
	defer b.lock.Unlock()

	key := client.Key()
	if c, ok := b.clients[key]; ok && c == client {
		delete(b.clients, key)

		// 更新用户平台映射
		b.updateUserPlatforms(client, false)

		// 更新IP统计
		if b.ipCount[client.IP()] > 1 {
			b.ipCount[client.IP()]--
		} else {
			delete(b.ipCount, client.IP())
		}
	}

	// 从房间中移除
	if client.room != nil {
		if client.room.DelClient(client) {
			// 房间为空，删除房间
			delete(b.rooms, client.room.ID)
		}
	}
}

// GetClient Get Client connection
func (b *Bucket) GetClient(key string) (*Client, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	client, ok := b.clients[key]
	return client, ok
}

// GetUserClients Retrieve all connections for the user
func (b *Bucket) GetUserClients(userID string) []*Client {
	b.lock.RLock()
	defer b.lock.RUnlock()

	userPlatforms, ok := b.userMap[userID]
	if !ok {
		return nil
	}

	userPlatforms.mutex.RLock()
	defer userPlatforms.mutex.RUnlock()

	var clients []*Client
	for _, platformClients := range userPlatforms.Platforms {
		clients = append(clients, platformClients...)
	}

	return clients
}

// GetUserPlatformClients Retrieve the user's specified platform connection
func (b *Bucket) GetUserPlatformClients(userID string, platformID int32) []*Client {
	b.lock.RLock()
	defer b.lock.RUnlock()

	userPlatforms, ok := b.userMap[userID]
	if !ok {
		return nil
	}

	userPlatforms.mutex.RLock()
	defer userPlatforms.mutex.RUnlock()

	return userPlatforms.Platforms[platformID]
}

// JoinRoom Join the room
func (b *Bucket) JoinRoom(client *Client, roomID, roomType string) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	// 离开当前房间
	if client.room != nil {
		if client.room.DelClient(client) {
			delete(b.rooms, client.room.ID)
		}
	}

	// 加入新房间
	room, ok := b.rooms[roomID]
	if !ok {
		room = NewRoom(roomID, roomType)
		b.rooms[roomID] = room
	}

	return room.AddClient(client)
}

// LeaveRoom Leave the room
func (b *Bucket) LeaveRoom(client *Client) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if client.room != nil {
		if client.room.DelClient(client) {
			delete(b.rooms, client.room.ID)
		}
	}
}

// GetRoom Get Room
func (b *Bucket) GetRoom(roomID string) (*Room, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	room, ok := b.rooms[roomID]
	return room, ok
}

// BroadcastRoom Room Broadcast
func (b *Bucket) BroadcastRoom(req *BroadcastReq) {
	idx := atomic.AddUint64(&b.routineNum, 1) % uint64(len(b.routines))
	select {
	case b.routines[idx] <- req:
	default:
		// 队列满时丢弃消息
	}
}

// Broadcast Global Broadcast
func (b *Bucket) Broadcast(msg []byte) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, client := range b.clients {
		select {
		case client.sendCh <- msg:
		default:
			// 发送队列满时跳过
		}
	}
}

// updateUserPlatforms Update User Platform Mapping
func (b *Bucket) updateUserPlatforms(client *Client, add bool) {
	userID := client.UserID
	platformID := client.PlatformID

	userPlatforms, ok := b.userMap[userID]
	if !ok {
		if !add {
			return
		}
		userPlatforms = &UserPlatforms{
			UserID:    userID,
			Platforms: make(map[int32][]*Client),
			lastTime:  time.Now().Unix(),
		}
		b.userMap[userID] = userPlatforms
	}

	userPlatforms.mutex.Lock()
	defer userPlatforms.mutex.Unlock()

	if add {
		userPlatforms.Platforms[platformID] = append(userPlatforms.Platforms[platformID], client)
		userPlatforms.lastTime = time.Now().Unix()
	} else {
		clients := userPlatforms.Platforms[platformID]
		for i, c := range clients {
			if c == client {
				userPlatforms.Platforms[platformID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		// 清理空平台
		if len(userPlatforms.Platforms[platformID]) == 0 {
			delete(userPlatforms.Platforms, platformID)
		}

		// 清理空用户
		if len(userPlatforms.Platforms) == 0 {
			delete(b.userMap, userID)
		}
	}
}

// roomProcessor Process Room Messages
func (b *Bucket) roomProcessor(ch chan *BroadcastReq) {
	for req := range ch {
		b.lock.RLock()
		room, ok := b.rooms[req.RoomID]
		b.lock.RUnlock()

		if ok {
			room.Broadcast(req.Message)
		}
	}
}
