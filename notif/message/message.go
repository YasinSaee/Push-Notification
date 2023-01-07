package message

import (
	"github.com/YasinSaee/Push-Notification/config"
	"github.com/YasinSaee/Push-Notification/context"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Message struct {
		ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
		Title       string             `bson:"title" json:"title"`
		Link        string             `bson:"link" json:"link"`
		Description string             `bson:"description" json:"description"`
		Image       string             `bson:"image" json:"image"`
		RoomId      primitive.ObjectID `bson:"room_id" json:"room_id"`
	}

	Messages []Message
)

const Collection = "message"

func (m *Message) Save() error {
	return config.MongoConn.Create(m)
}

func (m *Message) LoadByID(id interface{}) error {
	return config.MongoConn.Get(Collection, id, m)
}

func (m *Messages) List(filter MessageFilter, pfilter context.PublicFilter) (int, int, error) {
	q := filter.GetFilters()
	totalCount, err := config.MongoConn.Count(Collection, q)
	if err != nil {
		log.Error("can not count message list error : ", err.Error())
	}

	totalPages := totalCount / pfilter.Limit
	if totalPages*pfilter.Limit < totalCount {
		totalPages++
	}

	return totalCount, totalPages, config.MongoConn.Find(Collection, q, m, pfilter.Limit, pfilter.Page, pfilter.Sort)
}
