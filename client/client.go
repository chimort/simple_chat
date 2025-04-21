package client

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	UserName string
}

func New(conn *websocket.Conn, username string) *Client {
	return &Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserName: username,
	}
}

func (c *Client) ReadPump(onMessage func([]byte)) {
	defer c.Conn.Close()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		onMessage(msg)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("ошибка записи в сокет:", err)
				return
			}
		}
	}
}
