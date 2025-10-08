package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazyfrankie/goim/interfaces/ws/compressor"
	wsctx "github.com/crazyfrankie/goim/interfaces/ws/context"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/go-playground/validator/v10"
)

type LongConnServer interface {
	Run(ctx context.Context) error
	wsHandler(w http.ResponseWriter, r *http.Request)
	GetUserAllCons(userID string) ([]*Client, bool)
	GetUserPlatformCons(userID string, platform int) ([]*Client, bool, bool)
	Validate(s any) error
	KickUserConn(client *Client) error
	UnRegister(c *Client)
	SetKickHandlerInfo(i *kickHandler)
	SubUserOnlineStatus(ctx context.Context, client *Client, data *Req) ([]byte, error)
	compressor.Compressor
	MessageHandler
}

type WebsocketServer struct {
	port              int
	wsMaxConnNum      int64
	registerChan      chan *Client
	unregisterChan    chan *Client
	kickHandlerChan   chan *kickHandler
	bucketManager     *BucketManager // 使用我们的BucketManager替代clients UserMap
	subscription      *Subscription
	clientPool        sync.Pool
	onlineUserNum     atomic.Int64
	onlineUserConnNum atomic.Int64
	handshakeTimeout  time.Duration
	writeBufferSize   int
	validate          *validator.Validate
	compressor.Compressor
	MessageHandler
}

type kickHandler struct {
	clientOK   bool
	oldClients []*Client
	newClient  *Client
}

func (ws *WebsocketServer) UnRegister(c *Client) {
	ws.unregisterChan <- c
}

func (ws *WebsocketServer) Validate(_ any) error {
	return nil
}

func (ws *WebsocketServer) GetUserAllCons(userID string) ([]*Client, bool) {
	var allClients []*Client
	buckets := ws.bucketManager.GetAllBuckets()

	for _, bucket := range buckets {
		clients := bucket.GetUserClients(userID)
		allClients = append(allClients, clients...)
	}

	return allClients, len(allClients) > 0
}

func (ws *WebsocketServer) GetUserPlatformCons(userID string, platform int) ([]*Client, bool, bool) {
	var allClients []*Client
	buckets := ws.bucketManager.GetAllBuckets()

	for _, bucket := range buckets {
		clients := bucket.GetUserPlatformClients(userID, int32(platform))
		allClients = append(allClients, clients...)
	}

	return allClients, len(allClients) > 0, len(allClients) > 0
}

func NewWebsocketServer(opts ...Option) *WebsocketServer {
	var config configs
	for _, o := range opts {
		o(&config)
	}

	v := validator.New()
	return &WebsocketServer{
		port:             0,
		wsMaxConnNum:     0,
		writeBufferSize:  0,
		handshakeTimeout: 0,
		clientPool: sync.Pool{
			New: func() any {
				return new(Client)
			},
		},
		registerChan:    make(chan *Client, 1000),
		unregisterChan:  make(chan *Client, 1000),
		kickHandlerChan: make(chan *kickHandler, 1000),
		validate:        v,
		bucketManager:   NewBucketManager(32, DefaultBucketConfig()), // 使用BucketManager
		subscription:    newSubscription(),
		Compressor:      compressor.NewCompressor(),
	}
}

func (ws *WebsocketServer) Run(ctx context.Context) error {
	var client *Client

	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case client = <-ws.registerChan:
				ws.registerClient(client)
			case client = <-ws.unregisterChan:
				ws.unregisterClient(client)
			case onlineInfo := <-ws.kickHandlerChan:
				ws.multiTerminalLoginChecker(onlineInfo.clientOK, onlineInfo.oldClients, onlineInfo.newClient)
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		wsSrv := http.Server{Addr: fmt.Sprintf(":%d", ws.port), Handler: nil}
		http.HandleFunc("/", ws.wsHandler)
		go func() {
			defer close(done)
			<-ctx.Done()
			_ = wsSrv.Shutdown(context.Background())
		}()
		err := wsSrv.ListenAndServe()
		if err == nil {
			err = fmt.Errorf("http server closed")
		}
		cancel(fmt.Errorf("msg gateway %w", err))
	}()

	<-ctx.Done()

	timeout := time.NewTimer(time.Second * 15)
	defer timeout.Stop()
	select {
	case <-timeout.C:
		logs.Warnf("msg gateway graceful stop timeout")
	case <-done:
		logs.Debugf("msg gateway graceful stop done")
	}
	return context.Cause(ctx)
}

func (ws *WebsocketServer) SetKickHandlerInfo(i *kickHandler) {
	ws.kickHandlerChan <- i
}

func (ws *WebsocketServer) registerClient(client *Client) {
	bucket := ws.bucketManager.GetBucket(client.UserID)
	err := bucket.PutClient(client)
	if err != nil {
		logs.Errorf("register client failed: %v", err)
		return
	}

	ws.onlineUserConnNum.Add(1)
	logs.Debugf("user online, userID: %s, platformID: %d, online user conn Num: %d",
		client.UserID, client.PlatformID, ws.onlineUserConnNum.Load())
}

func (ws *WebsocketServer) KickUserConn(client *Client) error {
	// 从BucketManager中移除客户端
	bucket := ws.bucketManager.GetBucket(client.UserID)
	bucket.DelClient(client)
	return client.KickOnlineMessage()
}

func (ws *WebsocketServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	connContext := wsctx.NewContext(w, r)

	if ws.onlineUserConnNum.Load() >= ws.wsMaxConnNum {
		httpError(connContext, fmt.Errorf("over max conn num limit"))
		return
	}

	err := connContext.ParseEssentialArgs()
	if err != nil {
		httpError(connContext, err)
		return
	}

	logs.Debugf("new conn, token: %s", connContext.GetToken())

	wsLongConn := newWebSocketConn(ws.handshakeTimeout, ws.writeBufferSize)
	if err := wsLongConn.GenerateConn(w, r); err != nil {
		logs.Warnf("long connection fails: %v", err)
		return
	}

	client := ws.clientPool.Get().(*Client)
	client.Reset(connContext, wsLongConn, ws)

	ws.registerChan <- client
	go client.Start()
}

func (ws *WebsocketServer) multiTerminalLoginChecker(clientOK bool, oldClients []*Client, newClient *Client) {
	//// 多终端登录检查逻辑，基本保持与open-im-server一致
	//// 这里简化实现，实际应该根据配置策略处理
	//if clientOK {
	//	for _, c := range oldClients {
	//		if c.token == newClient.token {
	//			continue // 相同token不踢出
	//		}
	//
	//		// 踢出旧连接
	//		bucket := ws.bucketManager.GetBucket(c.UserID)
	//		bucket.DelClient(c)
	//		err := c.KickOnlineMessage()
	//		if err != nil {
	//			logs.Warnf("KickOnlineMessage failed: %v", err)
	//		}
	//	}
	//}
}

func (ws *WebsocketServer) unregisterClient(client *Client) {
	defer ws.clientPool.Put(client)

	// 从BucketManager中移除客户端
	bucket := ws.bucketManager.GetBucket(client.UserID)
	bucket.DelClient(client)

	ws.onlineUserConnNum.Add(-1)
	ws.subscription.DelClient(client)

	logs.Debugf("user offline, userID: %s, close reason: %v, online user conn Num: %d",
		client.UserID, client.closedErr, ws.onlineUserConnNum.Load())
}

func getRemoteAdders(client []*Client) string {
	var ret string
	for i, c := range client {
		if i == 0 {
			ret = c.ctx.GetRemoteAddr()
		} else {
			ret += "@" + c.ctx.GetRemoteAddr()
		}
	}
	return ret
}
