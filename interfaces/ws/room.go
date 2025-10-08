package ws

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/crazyfrankie/goim/interfaces/ws/types"
)

var (
	ErrRoomDropped = errors.New("room has been dropped")
)

type Room struct {
	ID       string
	Type     string
	head     *Client
	lock     sync.RWMutex
	drop     bool
	online   int32
	metadata map[string]interface{}
}

func NewRoom(id, roomType string) *Room {
	return &Room{
		ID:       id,
		Type:     roomType,
		metadata: make(map[string]interface{}),
	}
}

// AddClient Add a client to the room
func (r *Room) AddClient(client *Client) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.drop {
		return ErrRoomDropped
	}

	// Insert into the head of the linked list
	if r.head != nil {
		r.head.Prev = client
	}
	client.Next = r.head
	client.Prev = nil
	r.head = client
	client.room = r

	atomic.AddInt32(&r.online, 1)
	return nil
}

// DelClient Remove the client from the room
func (r *Room) DelClient(client *Client) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	// Remove from the linked list
	if client.Next != nil {
		client.Next.Prev = client.Prev
	}
	if client.Prev != nil {
		client.Prev.Next = client.Next
	} else {
		r.head = client.Next
	}

	client.Next = nil
	client.Prev = nil
	client.room = nil

	atomic.AddInt32(&r.online, -1)

	// Check whether the room needs to be deleted.
	r.drop = r.online == 0
	return r.drop
}

// Broadcast Room Broadcast Message
func (r *Room) Broadcast(msg *types.Message) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for client := r.head; client != nil; client = client.Next {
		select {
		case client.sendCh <- msg:
		default:
			// Skip when the send queue is full
		}
	}
}

// BroadcastFilter Broadcast with filter
func (r *Room) BroadcastFilter(msg *types.Message, filter func(*Client) bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for client := r.head; client != nil; client = client.Next {
		if filter != nil && !filter(client) {
			continue
		}

		select {
		case client.sendCh <- msg:
		default:
			// Skip when the send queue is full
		}
	}
}

// GetClients Retrieve all clients in the room
func (r *Room) GetClients() []*Client {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var clients []*Client
	for client := r.head; client != nil; client = client.Next {
		clients = append(clients, client)
	}

	return clients
}

// GetOnlineCount Get the number of online users
func (r *Room) GetOnlineCount() int32 {
	return atomic.LoadInt32(&r.online)
}

// SetMetadata Set Room Metadata
func (r *Room) SetMetadata(key string, value interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.metadata[key] = value
}

// GetMetadata Retrieve room metadata
func (r *Room) GetMetadata(key string) (interface{}, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	value, ok := r.metadata[key]
	return value, ok
}

func (r *Room) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.drop = true

	// Close all client connections
	for client := r.head; client != nil; client = client.Next {
		client.close()
	}
}

// IsEmpty Check if the room is empty
func (r *Room) IsEmpty() bool {
	return atomic.LoadInt32(&r.online) == 0
}
