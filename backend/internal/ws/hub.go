package ws

import (
	"fmt"
	"sync"
)

var mu sync.Mutex

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			fmt.Println("Client registered: ", client)

			hub.clients[client.UserID] = client

		case client := <-hub.unregister:
			fmt.Println("Client unreg: ", client)
			if _, ok := hub.clients[client.UserID]; ok {
				delete(hub.clients, client.UserID)
				close(client.Send)
			}
		case event := <-hub.broadcast:
			fmt.Println("Event: ", event)
			for _, v := range hub.clients {
				v.Send <- event
			}
		}
	}
}

func (h *Hub) SendToUser(userID string, event Event) {
	if client, ok := h.clients[userID]; ok {
		client.Send <- event
	}
}
