package ws

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazyfrankie/goim/interfaces/ws/compressor"
	wsctx "github.com/crazyfrankie/goim/interfaces/ws/context"
	"github.com/crazyfrankie/goim/interfaces/ws/encoding"
	"github.com/crazyfrankie/goim/interfaces/ws/types"
	"github.com/crazyfrankie/goim/pkg/ctxcache"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/safego"
	"github.com/crazyfrankie/goim/pkg/sonic"
	"github.com/crazyfrankie/goim/types/consts"
)

const (
	MessageText   = 1
	MessageBinary = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 51200
)

type PingPongHandler func(string) error

type ClientConfig struct {
	SendQueueSize int
	RecvRingSize  int
	WriteTimeout  time.Duration
	ReadTimeout   time.Duration
}

type Client struct {
	conn Conn
	ctx  *wsctx.Context

	UserID       string
	PlatformID   int32
	Token        string
	SDKType      string
	IsBackground bool
	ConnID       string

	sendCh   chan *types.Message
	recvRing *Ring

	subscriptions map[string]struct{}
	subLock       sync.RWMutex
	encoder       encoding.Encoder
	compressor    compressor.Compressor

	closed     atomic.Bool
	closedErr  error
	lastActive int64

	room *Room
	Next *Client
	Prev *Client

	clientCtx context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	writeMutex sync.Mutex
}

func NewClient(conn Conn, ctx *wsctx.Context, config *ClientConfig) *Client {
	clientCtx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:          conn,
		ctx:           ctx,
		UserID:        ctx.GetUserID(),
		PlatformID:    stringToInt(ctx.GetPlatformID()),
		Token:         ctx.GetToken(),
		SDKType:       ctx.GetSDKType(),
		ConnID:        ctx.GetConnID(),
		sendCh:        make(chan *types.Message, config.SendQueueSize),
		recvRing:      NewRing(config.RecvRingSize),
		subscriptions: make(map[string]struct{}),
		encoder:       encoding.NewJSONEncoder(),
		compressor:    compressor.NewCompressor(),
		clientCtx:     clientCtx,
		cancel:        cancel,
		lastActive:    time.Now().Unix(),
	}

	return client
}

func (c *Client) Reset() {
	c.conn = nil
	c.ctx = nil
	c.UserID = ""
	c.PlatformID = 0
	c.Token = ""
	c.ConnID = ""
	c.closed.Store(false)
	c.closedErr = nil
	c.room = nil
	c.Next = nil
	c.Prev = nil

	c.subLock.Lock()
	for k := range c.subscriptions {
		delete(c.subscriptions, k)
	}
	c.subLock.Unlock()

	if c.recvRing != nil {
		c.recvRing.Reset()
	}
}

func (c *Client) Key() string {
	return c.ConnID
}

func (c *Client) IP() string {
	return c.ctx.GetRemoteAddr()
}

func (c *Client) readMessage() {
	defer func() {
		if r := recover(); r != nil {
			c.closedErr = safego.NewPanicErr(r, debug.Stack())
		}
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPingHandler(c.pingHandler)
	c.conn.SetPongHandler(c.pongHandler)
	_ = c.conn.SetReadDeadline(pongWait)
	c.activeHeartBeat(c.clientCtx)

	for {
		messageType, message, returnErr := c.conn.ReadMessage()
		if returnErr != nil {
			logs.CtxWarnf(c.ctx, "readMessage, err: %v, messageType: %d", returnErr, messageType)
			c.closedErr = returnErr
			return
		}

		logs.CtxInfof(c.ctx, "readMessage,messageType: %d", messageType)

		if c.closed.Load() {
			// The scenario where the connection has just been closed, but the coroutine has not exited
			c.closedErr = types.ErrConnClosed
			return
		}

		switch messageType {
		case MessageBinary:
			_ = c.conn.SetReadDeadline(pongWait)
			parseDataErr := c.handleMessage(message)
			if parseDataErr != nil {
				c.closedErr = parseDataErr
				return
			}
		case MessageText:
			_ = c.conn.SetReadDeadline(pongWait)
			parseDataErr := c.handleTextMessage(message)
			if parseDataErr != nil {
				c.closedErr = parseDataErr
				return
			}
		case PingMessage:
			err := c.writePongMsg("")
			logs.CtxErrorf(c.ctx, "writePongMsg err: %v", err)

		case CloseMessage:
			c.closedErr = types.ErrClientClosed
			return

		default:
		}
	}
}

func (c *Client) handleMessage(message []byte) error {
	ctx := ctxcache.Init(context.Background())

	var binaryReq = getReq()
	defer freeReq(binaryReq)

	err := c.encoder.Decode(message, binaryReq)
	if err != nil {
		return err
	}

	//if err := c.longConnServer.Validate(binaryReq); err != nil {
	//	return err
	//}

	if binaryReq.SendID != c.UserID {
		return errorx.Wrapf(nil, "exception conn userID not same to req userID, binaryReq: %s", binaryReq.String())
	}

	ctxcache.StoreM(ctx, binaryReq.OperationID, binaryReq.SendID, consts.PlatformIDToName(c.PlatformID), c.ctx.GetConnID())

	logs.CtxDebugf(ctx, "gateway req message, req: %s", binaryReq.String())

	var (
		resp       []byte
		messageErr error
	)

	switch binaryReq.ReqIdentifier {
	case types.WSGetNewestSeq:
		//resp, messageErr = c.longConnServer.GetSeq(ctx, binaryReq)
	case types.WSSendMsg:
		//resp, messageErr = c.longConnServer.SendMessage(ctx, binaryReq)
	case types.WSSendSignalMsg:
		//resp, messageErr = c.longConnServer.SendSignalMessage(ctx, binaryReq)
	case types.WSPullMsgBySeqList:
		//resp, messageErr = c.longConnServer.PullMessageBySeqList(ctx, binaryReq)
	case types.WSPullMsg:
		//resp, messageErr = c.longConnServer.GetSeqMessage(ctx, binaryReq)
	case types.WSGetConvMaxReadSeq:
		//resp, messageErr = c.longConnServer.GetConversationsHasReadAndMaxSeq(ctx, binaryReq)
	case types.WsPullConvLastMessage:
		//resp, messageErr = c.longConnServer.GetLastMessage(ctx, binaryReq)
	case types.WsLogoutMsg:
		//resp, messageErr = c.longConnServer.UserLogout(ctx, binaryReq)
	case types.WsSetBackgroundStatus:
		//resp, messageErr = c.setAppBackgroundStatus(ctx, binaryReq)
	case types.WsSubUserOnlineStatus:
		//resp, messageErr = c.longConnServer.SubUserOnlineStatus(ctx, c, binaryReq)
	default:
		return fmt.Errorf(
			"ReqIdentifier failed,sendID:%s,msgIncr:%s,reqIdentifier:%d",
			binaryReq.SendID,
			binaryReq.MsgIncr,
			binaryReq.ReqIdentifier,
		)
	}

	return c.replyMessage(ctx, binaryReq, messageErr, resp)
}

func (c *Client) handleTextMessage(message []byte) error {
	var msg TextMessage
	if err := sonic.Unmarshal(message, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case TextPong:
		return nil
	case TextPing:
		msg.Type = TextPong
		msgData, err := sonic.Marshal(msg)
		if err != nil {
			return err
		}
		c.writeMutex.Lock()
		defer c.writeMutex.Unlock()
		if err := c.conn.SetWriteDeadline(writeWait); err != nil {
			return err
		}
		return c.conn.WriteMessage(MessageText, msgData)
	default:
		return fmt.Errorf("not support message type %s", msg.Type)
	}
}

func (c *Client) replyMessage(ctx context.Context, binaryReq *Req, err error, resp []byte) error {
	errResp := response.ParseError(err)
	mReply := &Resp{
		ReqIdentifier: binaryReq.ReqIdentifier,
		MsgIncr:       binaryReq.MsgIncr,
		OperationID:   binaryReq.OperationID,
		ErrCode:       errResp.Code,
		ErrMsg:        errResp.Message,
		Data:          resp,
	}
	t := time.Now()
	logs.CtxDebugf(ctx, "gateway reply message, resp: %s", mReply.String())
	err = c.writeBinaryMsg(mReply)
	if err != nil {
		logs.CtxWarnf(ctx, "wireBinaryMsg replyMessage, err: %v, resp: %s", err, mReply.String())
	}
	logs.CtxDebugf(ctx, "wireBinaryMsg end, time cost: %v", time.Since(t))

	if binaryReq.ReqIdentifier == types.WsLogoutMsg {
		return errorx.Wrapf(nil, "operationID, %s", binaryReq.OperationID)
	}

	return nil
}

func (c *Client) writeBinaryMsg(resp *Resp) error {
	if c.closed.Load() {
		return types.ErrClientClosed
	}

	encodedBuf, err := c.encoder.Encode(resp)
	if err != nil {
		return err
	}

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	err = c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	if c.ctx.GetCompression() && len(encodedBuf) > 1024 {
		compressed, err := c.compressor.Compress(encodedBuf)
		if err == nil {
			encodedBuf = compressed
		}
	}

	return c.conn.WriteMessage(MessageBinary, encodedBuf)
}

func (c *Client) close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	c.cancel()
	if c.sendCh != nil {
		close(c.sendCh)
	}

	err := c.conn.Close()
	c.wg.Wait()

	return err
}

func (c *Client) activeHeartBeat(ctx context.Context) {
	safego.Go(ctx, func() {
		logs.CtxDebugf(ctx, "server initiative send heartbeat start.")
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.writePingMsg(); err != nil {
					logs.CtxWarnf(c.ctx, "send Ping Message error, %v", err)
					return
				}
			case <-c.clientCtx.Done():
				return
			}
		}
	})
}

func (c *Client) pingHandler(appData string) error {
	if err := c.conn.SetReadDeadline(pongWait); err != nil {
		return err
	}

	logs.CtxDebugf(c.ctx, "ping Handler Success, appData, %v", appData)
	return c.writePongMsg(appData)
}

func (c *Client) pongHandler(_ string) error {
	if err := c.conn.SetReadDeadline(pongWait); err != nil {
		return err
	}
	return nil
}

func (c *Client) writePingMsg() error {
	if c.closed.Load() {
		return nil
	}

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(PingMessage, nil)
}

func (c *Client) writePongMsg(appData string) error {
	logs.CtxDebugf(c.ctx, "write Pong Msg in Server, appData: %v", appData)
	if c.closed.Load() {
		logs.CtxWarnf(c.ctx, "is closed in server, appdata: %v, closed err: %v", appData, c.closedErr)
		return nil
	}

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		logs.CtxWarnf(c.ctx, "SetWriteDeadline in Server have error, %v, writeWait: %d, appData: %v", err, writeWait, appData)
		return errorx.Wrapf(err, "")
	}
	err = c.conn.WriteMessage(PongMessage, []byte(appData))
	if err != nil {
		logs.CtxWarnf(c.ctx, "Write Message have error, %v, Pong msg: %d", err, PongMessage)
	}

	return errorx.Wrapf(err, "")
}

func stringToInt(s string) int32 {
	var result int32
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + char - '0'
		}
	}
	return result
}
