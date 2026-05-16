package ws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	Send   chan Event
	UserID string
	Hub    *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var event Event
		if err = json.Unmarshal(message, &event); err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		c.Hub.broadcast <- event
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case event, ok := <-c.Send:
			fmt.Println(event, ok)
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteJSON(event)
		}
	}
}
