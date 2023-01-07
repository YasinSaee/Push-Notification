package notif

import (
	"github.com/YasinSaee/Push-Notification/notif/message"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Hub struct {
		Rooms      map[primitive.ObjectID]*Room
		Register   chan *Client
		Unregister chan *Client
		Broadcast  chan *message.Message
	}
)

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[primitive.ObjectID]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *message.Message, 5),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			{
				if _, ok := h.Rooms[cl.RoomId]; ok {
					r := h.Rooms[cl.RoomId]
					if _, ok := r.Cleints[cl.Customer]; !ok {
						r.Cleints[cl.Customer] = cl
					}
				}
			}
		case cl := <-h.Unregister:
			{
				if _, ok := h.Rooms[cl.RoomId]; ok {
					if _, ok := h.Rooms[cl.RoomId].Cleints[cl.Customer]; ok {
						if len(h.Rooms[cl.RoomId].Cleints) != 0 {

						}

						delete(h.Rooms[cl.RoomId].Cleints, cl.Customer)
						close(cl.Message)
					}
				}
			}
		case m := <-h.Broadcast:
			{
				if _, ok := h.Rooms[m.RoomId]; ok {
					for _, cl := range h.Rooms[m.RoomId].Cleints {
						cl.Message <- m
					}
				}
			}

		}
	}
}
