package notif

import (
	"github.com/YasinSaee/Push-Notification/config"
	"github.com/YasinSaee/Push-Notification/context"
	"github.com/YasinSaee/Push-Notification/notif/client"
	"github.com/YasinSaee/Push-Notification/notif/message"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Client struct {
		Conn     *websocket.Conn       `bson:"conn" json:"conn"`
		Message  chan *message.Message `bson:"message" json:"message"`
		Customer primitive.ObjectID    `bson:"customer" json:"customer"`
		RoomId   primitive.ObjectID    `bson:"room_id" json:"room_id"`
		Username string                `bson:"username" json:"username"`
	}
	ClientRes struct {
		ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
		Conn     *websocket.Conn    `bson:"conn" json:"conn"`
		Customer primitive.ObjectID `bson:"customer" json:"customer"`
		RoomId   primitive.ObjectID `bson:"room_id" json:"room_id"`
		Username string             `bson:"username" json:"username"`
	}
	Clients []ClientRes
)

const Collection = "client_res"

func (c *Client) readMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}

}

func (c *ClientRes) LoadByID(roomId interface{}, custId primitive.ObjectID) error {
	var (
		objID primitive.ObjectID
		err   error
	)

	if val, ok := roomId.(string); ok {
		objID, err = primitive.ObjectIDFromHex(val)
		if err != nil {
			return err
		}
	}
	return config.MongoConn.FindOne(Collection, bson.M{"room_id": objID, "customer": custId}, c)
}

func (c *ClientRes) Save() error {
	if !c.ID.IsZero() {
		return config.MongoConn.Update(c)
	}
	return config.MongoConn.Create(c)
}

func (c *Clients) List(filter client.ClientFilter, pfilter context.PublicFilter) (int, int, error) {
	q := filter.GetFilters()
	totalCount, err := config.MongoConn.Count(Collection, q)
	if err != nil {
		log.Error("can not count music list error : ", err.Error())
	}

	totalPages := totalCount / pfilter.Limit
	if totalPages*pfilter.Limit < totalCount {
		totalPages++
	}

	return totalCount, totalPages, config.MongoConn.Find(Collection, q, c, pfilter.Limit, pfilter.Page, pfilter.Sort)
}
