package ws

import (
	"sync"

	"github.com/crazyfrankie/goim/interfaces/ws/types"
)

type RoundConfig struct {
	ReaderNum     int
	ReaderBuf     int
	ReaderBufSize int
	WriterNum     int
	WriterBuf     int
	WriterBufSize int
	TimerNum      int
	TimerSize     int
}

// Ring Ring Buffer
type Ring struct {
	// Read/Write Pointers (Using Different Cache Lines to Avoid False Sharing)
	rp uint64
	_  [7]uint64 // Cache Line Padding
	wp uint64
	_  [7]uint64 // Cache Line Padding

	// Buffer
	mask uint64
	data []types.Message
}

func NewRing(size int) *Ring {
	// 确保size是2的幂
	if size&(size-1) != 0 {
		for size&(size-1) != 0 {
			size &= size - 1
		}
		size <<= 1
	}

	return &Ring{
		mask: uint64(size - 1),
		data: make([]types.Message, size),
	}
}

// Set Obtain the write position
func (r *Ring) Set() (*types.Message, error) {
	if r.wp-r.rp >= uint64(len(r.data)) {
		return nil, types.ErrRingFull
	}

	return &r.data[r.wp&r.mask], nil
}

// SetAdv Advance the pointer
func (r *Ring) SetAdv() {
	r.wp++
}

// Get Returns the reading position
func (r *Ring) Get() (*types.Message, error) {
	if r.rp == r.wp {
		return nil, types.ErrRingEmpty
	}

	return &r.data[r.rp&r.mask], nil
}

// GetAdv Advance the pointer
func (r *Ring) GetAdv() {
	r.rp++
}

// Reset Reset Buffer
func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}

// Len Get current length
func (r *Ring) Len() int {
	return int(r.wp - r.rp)
}

func (r *Ring) Cap() int {
	return len(r.data)
}

// Round Resource Pool
type Round struct {
	readers []sync.Pool
	writers []sync.Pool

	readerNum int
	writerNum int
}

func NewRound(config *RoundConfig) *Round {
	r := &Round{
		readers:   make([]sync.Pool, config.ReaderNum),
		writers:   make([]sync.Pool, config.WriterNum),
		readerNum: config.ReaderNum,
		writerNum: config.WriterNum,
	}

	// 初始化读缓冲池
	for i := 0; i < config.ReaderNum; i++ {
		r.readers[i].New = func() interface{} {
			return make([]byte, config.ReaderBufSize)
		}
	}

	// 初始化写缓冲池
	for i := 0; i < config.WriterNum; i++ {
		r.writers[i].New = func() interface{} {
			return make([]byte, config.WriterBufSize)
		}
	}

	return r
}

// GetReader Acquire read buffer
func (r *Round) GetReader(n int) []byte {
	return r.readers[n%r.readerNum].Get().([]byte)
}

// PutReader Return read buffer
func (r *Round) PutReader(n int, buf []byte) {
	r.readers[n%r.readerNum].Put(buf[:0]) // Reset length while preserving capacity
}

// GetWriter Acquire write buffer
func (r *Round) GetWriter(n int) []byte {
	return r.writers[n%r.writerNum].Get().([]byte)
}

// PutWriter Return write buffer
func (r *Round) PutWriter(n int, buf []byte) {
	r.writers[n%r.writerNum].Put(buf[:0]) // Reset length while preserving capacity
}
