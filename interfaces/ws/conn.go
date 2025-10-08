package ws

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/crazyfrankie/goim/pkg/errorx"
)

type Conn interface {
	// Close this connection
	Close() error
	// WriteMessage write message to connection,messageType means data type
	WriteMessage(messageType int, data []byte) error
	// ReadMessage read message from this connection
	ReadMessage() (int, []byte, error)
	// SetReadDeadline sets the read deadline on the underlying network connection,
	// after a read has timed out, will return an error.
	SetReadDeadline(timeout time.Duration) error
	// SetWriteDeadline sets to write deadline when send message,when read has timed out,will return error.
	SetWriteDeadline(timeout time.Duration) error
	// GenerateConn Check the connection of the current and when it was sent are the same
	GenerateConn(w http.ResponseWriter, r *http.Request) error
	// SetReadLimit sets the maximum size for a message read from the peer.bytes
	SetReadLimit(limit int64)
	SetPongHandler(handler PingPongHandler)
	SetPingHandler(handler PingPongHandler)
}

type WebSocketConn struct {
	conn             *websocket.Conn
	writeBufferSize  int
	handshakeTimeout time.Duration
}

func newWebSocketConn(handshakeTimeout time.Duration, wbs int) *WebSocketConn {
	return &WebSocketConn{handshakeTimeout: handshakeTimeout, writeBufferSize: wbs}
}

func (wc *WebSocketConn) Close() error {
	return wc.conn.Close()
}

func (wc *WebSocketConn) WriteMessage(messageType int, data []byte) error {
	return wc.conn.WriteMessage(messageType, data)
}

func (wc *WebSocketConn) ReadMessage() (int, []byte, error) {
	return wc.conn.ReadMessage()
}

func (wc *WebSocketConn) SetReadDeadline(timeout time.Duration) error {
	return wc.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (wc *WebSocketConn) SetWriteDeadline(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	// TODO SetWriteDeadline Future add error handling
	if err := wc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return errorx.Wrapf(err, "WebSocketConn.SetWriteDeadline failed")
	}

	return nil
}

func (wc *WebSocketConn) SetReadLimit(limit int64) {
	wc.conn.SetReadLimit(limit)
}

func (wc *WebSocketConn) SetPongHandler(handler PingPongHandler) {
	wc.conn.SetPongHandler(handler)
}

func (wc *WebSocketConn) SetPingHandler(handler PingPongHandler) {
	wc.conn.SetPingHandler(handler)
}

func (wc *WebSocketConn) GenerateConn(w http.ResponseWriter, r *http.Request) error {
	upgrader := &websocket.Upgrader{
		HandshakeTimeout: wc.handshakeTimeout,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	if wc.writeBufferSize > 0 { // default is 4kb.
		upgrader.WriteBufferSize = wc.writeBufferSize
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// The upgrader.Upgrade method usually returns enough error messages to diagnose problems that may occur during the upgrade
		return errorx.Wrapf(err, "GenerateConn: WebSocket upgrade failed")
	}
	wc.conn = conn
	return nil
}
