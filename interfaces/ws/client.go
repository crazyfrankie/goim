package ws

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	wsctx "github.com/crazyfrankie/goim/interfaces/ws/context"
	"github.com/crazyfrankie/goim/interfaces/ws/encoding"
	"github.com/crazyfrankie/goim/interfaces/ws/types"
	"github.com/crazyfrankie/goim/pkg/ctxcache"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/safego"
	"github.com/crazyfrankie/goim/pkg/sonic"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

const (
	// MessageText is for UTF-8 encoded text messages like JSON.
	MessageText = iota + 1
	// MessageBinary is for binary messages like protobufs.
	MessageBinary
	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
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
	ConnID       string
	IsBackground bool
	IsCompress   bool

	sendCh   chan []byte
	recvRing *Ring

	subscriptions map[string]struct{}
	subLock       sync.RWMutex

	encoder encoding.Encoder

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

	ConnServer LongConnServer
}

func NewClient(conn Conn, ctx *wsctx.Context, config *ClientConfig, connServer LongConnServer) *Client {
	clientCtx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:          conn,
		ctx:           ctx,
		UserID:        ctx.GetUserID(),
		PlatformID:    stringToInt(ctx.GetPlatformID()),
		Token:         ctx.GetToken(),
		SDKType:       ctx.GetSDKType(),
		ConnID:        ctx.GetConnID(),
		sendCh:        make(chan []byte, config.SendQueueSize),
		recvRing:      NewRing(config.RecvRingSize),
		subscriptions: make(map[string]struct{}),
		encoder:       encoding.NewJSONEncoder(),
		clientCtx:     clientCtx,
		cancel:        cancel,
		lastActive:    time.Now().Unix(),
		ConnServer:    connServer,
	}

	return client
}

func (c *Client) Reset(ctx *wsctx.Context, conn Conn, wsSrv LongConnServer) {
	c.conn = conn
	c.ctx = ctx

	c.UserID = ctx.GetUserID()
	c.PlatformID = stringToInt(ctx.GetPlatformID())
	c.Token = ctx.GetToken()
	c.SDKType = ctx.GetSDKType()
	c.ConnID = ctx.GetConnID()
	c.IsCompress = ctx.GetCompression()
	c.IsBackground = false

	c.closed.Store(false)
	c.closedErr = nil
	c.lastActive = 0

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

	if c.SDKType == types.GoSDK {
		c.encoder = encoding.NewGobEncoder()
	} else {
		c.encoder = encoding.NewJSONEncoder()
	}

	if c.sendCh != nil {
		for {
			select {
			case <-c.sendCh:
			default:
				goto done
			}
		}
	done:
	}

	c.ConnServer = wsSrv

	// 取消上下文
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.clientCtx = nil
}

func (c *Client) Key() string {
	return c.ConnID
}

func (c *Client) IP() string {
	return c.ctx.GetRemoteAddr()
}

func (c *Client) Start() {
	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) KickOnlineMessage() error {
	resp := &Resp{
		ReqIdentifier: types.WSKickOnlineMsg,
	}
	logs.CtxDebugf(c.ctx, "KickOnlineMessage debug")
	err := c.sendResp(resp)
	c.close()
	return err
}

func (c *Client) PushUserOnlineStatus(data []byte) error {
	resp := &Resp{
		ReqIdentifier: types.WsSubUserOnlineStatus,
		Data:          data,
	}
	return c.sendResp(resp)
}

func (c *Client) PushMessage(ctx context.Context, msgData *messagev1.Message) error {
	//var msg *messagev1.Message
	//conversationID := msgprocessor.GetConversationIDByMsg(msgData)
	//m := map[string]*sdkws.PullMsgs{conversationID: {Msgs: []*sdkws.MsgData{msgData}}}
	//if msgprocessor.IsNotification(conversationID) {
	//	msg.NotificationMsgs = m
	//} else {
	//	msg.Msgs = m
	//}
	//log.ZDebug(ctx, "PushMessage", "msg", &msg)
	//data, err := proto.Marshal(&msg)
	//if err != nil {
	//	return err
	//}
	//resp := Resp{
	//	ReqIdentifier: WSPushMsg,
	//	OperationID:   mcontext.GetOperationID(ctx),
	//	Data:          data,
	//}
	//return c.writeBinaryMsg(resp)
	return nil
}

func (c *Client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.closedErr = safego.NewPanicErr(r, debug.Stack())
		}
		c.close()
		c.wg.Done()
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
			c.closedErr = types.ErrConnClosed
			return
		}

		c.lastActive = time.Now().Unix()

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

// writeLoop 写入消息循环
func (c *Client) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			logs.CtxErrorf(c.ctx, "writeLoop panic: %v", r)
		}
		c.wg.Done()
	}()

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	var batch [][]byte
	const maxBatchSize = 10

	for {
		select {
		case msgData, ok := <-c.sendCh:
			if !ok {
				if len(batch) > 0 {
					c.flushBatch(batch)
				}
				return
			}

			batch = append(batch, msgData)
			if len(batch) >= maxBatchSize {
				c.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				c.flushBatch(batch)
				batch = batch[:0]
			}

		case <-c.clientCtx.Done():
			if len(batch) > 0 {
				c.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch 批量发送消息
func (c *Client) flushBatch(batch [][]byte) {
	for _, msgData := range batch {
		if err := c.writeMessage(msgData); err != nil {
			logs.CtxWarnf(c.ctx, "writeRawMessage failed: %v", err)
			c.close()
			return
		}
	}
}

func (c *Client) handleMessage(message []byte) error {
	if c.IsCompress && c.ConnServer != nil {
		if decompressed, err := c.ConnServer.Decompress(message); err == nil {
			message = decompressed
		}
	}

	return c.processMessage(message)
}

func (c *Client) processMessage(message []byte) error {
	ctx := ctxcache.Init(context.Background())

	var binaryReq = getReq()
	defer freeReq(binaryReq)

	err := c.encoder.Decode(message, binaryReq)
	if err != nil {
		return err
	}

	if err := c.ConnServer.Validate(binaryReq); err != nil {
		return err
	}

	if binaryReq.SendID != c.UserID {
		return errorx.Wrapf(nil, "exception conn userID not same to req userID, binaryReq: %s", binaryReq.String())
	}

	ctxcache.StoreM(ctx,
		types.OperationID, binaryReq.OperationID,
		types.WsUserID, binaryReq.SendID,
		types.PlatformID, consts.PlatformIDToName(c.PlatformID),
		types.ConnID, c.ctx.GetConnID())

	logs.CtxDebugf(ctx, "gateway req message, req: %s", binaryReq.String())

	var (
		resp       []byte
		messageErr error
	)

	switch binaryReq.ReqIdentifier {
	case types.WSGetNewestSeq:
		resp, messageErr = c.ConnServer.GetSeq(ctx, binaryReq)
	case types.WSSendMsg:
		resp, messageErr = c.ConnServer.SendMessage(ctx, binaryReq)
	case types.WSSendSignalMsg:
		resp, messageErr = c.ConnServer.SendSignalMessage(ctx, binaryReq)
	case types.WSPullMsgBySeqList:
		resp, messageErr = c.ConnServer.PullMessageBySeqList(ctx, binaryReq)
	case types.WSPullMsg:
		resp, messageErr = c.ConnServer.GetSeqMessage(ctx, binaryReq)
	case types.WSGetConvMaxReadSeq:
		resp, messageErr = c.ConnServer.GetConversationsHasReadAndMaxSeq(ctx, binaryReq)
	case types.WsPullConvLastMessage:
		resp, messageErr = c.ConnServer.GetLastMessage(ctx, binaryReq)
	case types.WsLogoutMsg:
		resp, messageErr = c.ConnServer.UserLogout(ctx, binaryReq)
	case types.WsSetBackgroundStatus:
		resp, messageErr = c.setAppBackgroundStatus(ctx, binaryReq)
	case types.WsSubUserOnlineStatus:
		resp, messageErr = c.ConnServer.SubUserOnlineStatus(ctx, c, binaryReq)
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
	err = c.sendResp(mReply)
	if err != nil {
		logs.CtxWarnf(ctx, "wireBinaryMsg replyMessage, err: %v, resp: %s", err, mReply.String())
	}
	logs.CtxDebugf(ctx, "wireBinaryMsg end, time cost: %v", time.Since(t))

	if binaryReq.ReqIdentifier == types.WsLogoutMsg {
		go func() {
			time.Sleep(time.Millisecond * 100)
			c.close()
		}()
		return errorx.Wrapf(nil, "user logout, operationID: %s", binaryReq.OperationID)
	}

	return nil
}

func (c *Client) close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	if c.ConnServer != nil {
		c.ConnServer.UnRegister(c)
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

func (c *Client) sendResp(resp *Resp) error {
	data, err := c.encoder.Encode(resp)
	if err != nil {
		return err
	}

	if c.closed.Load() {
		return types.ErrClientClosed
	}

	select {
	case c.sendCh <- data:
		return nil
	default:
		return types.ErrSendQueueFull
	}
}

func (c *Client) writeMessage(data []byte) error {
	if c.closed.Load() {
		return types.ErrClientClosed
	}

	originalSize := len(data)

	if c.IsCompress && len(data) > 1024 {
		if compressed, err := c.ConnServer.Compress(data); err == nil {
			data = compressed
			logs.CtxDebugf(c.ctx, "message compressed: %d -> %d bytes (%.1f%%)",
				originalSize, len(data),
				float64(len(data))/float64(originalSize)*100)
		}
	}

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if err := c.conn.SetWriteDeadline(writeWait); err != nil {
		return err
	}

	if err := c.conn.WriteMessage(MessageBinary, data); err != nil {
		return err
	}

	c.lastActive = time.Now().Unix()

	return nil
}

func (c *Client) KickOut() error {
	if c.ConnServer != nil {
		return c.ConnServer.KickUserConn(c)
	}
	return c.close()
}

func (c *Client) setAppBackgroundStatus(ctx context.Context, binaryReq *Req) ([]byte, error) {
	resp, isBackground, messageErr := c.ConnServer.SetUserDeviceBackground(ctx, binaryReq)
	if messageErr != nil {
		return nil, messageErr
	}

	c.IsBackground = isBackground
	// TODO: callback
	return resp, nil
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
