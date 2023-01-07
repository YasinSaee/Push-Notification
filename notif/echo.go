package notif

import (
	"net/http"
	"notification/context"
	"notification/notif/client"
	"notification/notif/customer_message"
	"notification/notif/message"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Handler struct {
		hub *Hub
	}
	CreateRoomReq struct {
		Name string `json:"name"`
	}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		/* 		origin := r.Header.Get("Origin")
		   		return origin == "http://aod.ir:3200" */
		return true
	},
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

func (h *Handler) CreateRoom(c echo.Context) error {
	var (
		req CreateRoomReq
		err error
	)
	g := c.(*context.GlobalContext)
	if err = g.Bind(&req); err != nil {
		return g.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	r := &Room{
		Name:    req.Name,
		Cleints: make(map[primitive.ObjectID]*Client),
	}

	if err = r.Save(); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	h.hub.Rooms[r.ID] = r

	return g.JSON(http.StatusOK, echo.Map{
		"room": r,
	})
}

func (h *Handler) GetRooms(c echo.Context) error {
	/* var (
		err error
	) */
	g := c.(*context.GlobalContext)

	rooms := make(Rooms, 0)
	/* if _, _, err = rooms.List(room.NotifFilter{}, context.PublicFilter{Limit: 110, Page: 1, Sort: ""}); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} */

	for _, r := range rooms {
		h.hub.Rooms[r.ID] = &r
	}

	return g.JSON(http.StatusOK, rooms)
}

func (h *Handler) GetMessages(c echo.Context) error {
	var (
		err error
	)
	g := c.(*context.GlobalContext)
	roomId := g.Param("roomId")

	messages := make(message.Messages, 0)
	if _, _, err = messages.List(message.MessageFilter{Room: roomId}, context.PublicFilter{Limit: 110, Page: 1, Sort: ""}); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return g.JSON(http.StatusOK, echo.Map{
		"messages": messages,
	})
}

func (h *Handler) GetCustomerMessages(c echo.Context) error {
	var (
		err error
	)
	g := c.(*context.GlobalContext)

	limit := g.QueryParam("limit")
	page := g.QueryParam("page")
	sort := g.QueryParam("sort")
	pfilter := context.SetFilter(limit, page, sort)

	key := g.QueryParam("key")
	messageId := g.QueryParam("message")

	/* cust := new(customer.Customer)
	if err = cust.CurrentUser(g); err != nil {
		log.Error("a problem has occurred. please try again in a few minutes, error : ", err.Error())
		return g.JSON(http.StatusInternalServerError, err.Error())
	} */
	c_ms := make(customer_message.CustomerMessages, 0)

	switch key {
	case "customer":
		{
			/* if _, _, err = c_ms.List(customer_message.CustomerMessageFilter{Customer: cust.ID.Hex()}, pfilter); err != nil {
				return g.JSON(http.StatusInternalServerError, err.Error())
			} */
		}
	case "msg":
		{
			if _, _, err = c_ms.List(customer_message.CustomerMessageFilter{Message: messageId}, pfilter); err != nil {
				return g.JSON(http.StatusInternalServerError, err.Error())
			}
		}
	}

	return g.JSON(http.StatusOK, echo.Map{
		"customer_messages": c_ms.Rest(),
	})
}

func (h *Handler) GetClients(c echo.Context) error {
	var (
		err error
	)

	g := c.(*context.GlobalContext)
	roomId := g.Param("roomId")

	clients := make(Clients, 0)
	if _, _, err = clients.List(client.ClientFilter{Room: roomId}, context.PublicFilter{Limit: 110, Page: 1, Sort: ""}); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	for _, c := range clients {
		cl := &Client{
			Conn:     c.Conn,
			Message:  make(chan *message.Message, 10),
			Customer: c.Customer,
			RoomId:   c.RoomId,
			Username: c.Username,
		}
		h.hub.Register <- cl
	}

	return g.JSON(http.StatusOK, echo.Map{
		"clients": clients,
	})
}

func (h *Handler) SendMessage(c echo.Context) error {
	var (
		err error
	)

	g := c.(*context.GlobalContext)
	roomId := g.Param("roomId")

	msg := new(message.Message)
	form := message.Message{}
	if err = g.Bind(&form); err != nil {
		return g.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	clients := make(Clients, 0)
	if _, _, err = clients.List(client.ClientFilter{Room: roomId}, context.PublicFilter{Limit: 110, Page: 1, Sort: ""}); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	msg.Description = form.Description
	msg.Title = form.Title
	msg.Image = form.Image
	msg.Link = form.Link
	msg.RoomId, _ = primitive.ObjectIDFromHex(roomId)

	if err = msg.Save(); err != nil {
		return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	for _, cl := range clients {
		c_m := new(customer_message.CustomerMessage)
		c_m.Customer = cl.Customer
		c_m.Message = msg.ID
		if err = c_m.Save(); err != nil {
			return g.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}

	h.hub.Broadcast <- msg

	return g.JSON(200, msg)
}

func (h *Handler) JoinRoom(c echo.Context) error {
	var (
		err error
	)

	g := c.(*context.GlobalContext)
	conn, err := upgrader.Upgrade(g.Response(), g.Request(), nil)
	if err != nil {
		log.Error("a problem has occurred. please try again in a few minutes, error : ", err.Error())
		return g.JSON(http.StatusInternalServerError, err.Error())
	}

	/* cust := new(customer.Customer)
	if err = cust.CurrentUser(g); err != nil {
		log.Error("a problem has occurred. please try again in a few minutes, error : ", err.Error())
		return g.JSON(http.StatusInternalServerError, "resp")
	} */

	roomId := g.Param("roomId")
	cl := new(Client)
	cl.Conn = conn
	cl.Message = make(chan *message.Message, 10)
	/* 	cl.Customer = cust.ID
	 */cl.RoomId, _ = primitive.ObjectIDFromHex(roomId)
	/* 	cl.Username = cust.Cellphone */

	//sclR := new(ClientRes)
	/* if err = clR.LoadByID(roomId, cust.ID); err != nil {
		clR.Conn = conn
		clR.Customer = cust.ID
		clR.Username = cust.Cellphone
		clR.RoomId, _ = primitive.ObjectIDFromHex(roomId)

		if err = clR.Save(); err != nil {
			return g.JSON(500, err.Error())
		}
	} */

	h.hub.Register <- cl
	go cl.readMessage()

	return g.JSON(200, nil)
}
