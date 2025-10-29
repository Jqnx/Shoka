package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	writeMux sync.Mutex
}

func (c *Client) WriteMessage(messageType int, data []byte) error {
	c.writeMux.Lock()
	defer c.writeMux.Unlock()
	return c.conn.WriteMessage(messageType, data)
}

func (c *Client) Close() error {
	c.writeMux.Lock()
	defer c.writeMux.Unlock()
	return c.conn.Close()
}
