package notif

import (
	"github.com/labstack/echo"
)

func Register(EP *echo.Echo) {
	hub := NewHub()
	wsHandler := NewHandler(hub)
	go hub.Run()

	notifGroup := EP.Group("/notif")
	notifGroup.POST("/ws/create/room", wsHandler.CreateRoom)
	notifGroup.POST("/ws/send/:roomId", wsHandler.SendMessage)
	notifGroup.GET("/ws/get/room", wsHandler.GetRooms)
	notifGroup.GET("/ws/get/clients/:roomId", wsHandler.GetClients)
	notifGroup.GET("/ws/get/messages/:roomId", wsHandler.GetMessages)
	notifGroup.GET("/ws/get/messages", wsHandler.GetCustomerMessages)
	notifGroup.GET("/ws/joinroom/:roomId", wsHandler.JoinRoom)

}
