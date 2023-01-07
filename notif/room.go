package notif

import (
	"github.com/YasinSaee/Push-Notification/config"
	"github.com/YasinSaee/Push-Notification/context"
	"github.com/YasinSaee/Push-Notification/notif/room"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Room struct {
		ID      primitive.ObjectID             `bson:"_id,omitempty" json:"_id"`
		Name    string                         `json:"name"`
		Cleints map[primitive.ObjectID]*Client `bson:"clients" json:"clients"`
	}
	Rooms []Room
)

const Collection_Room = "room"

func (r *Room) Save() error {
	if !r.ID.IsZero() {
		return config.MongoConn.Update(r)
	}

	return config.MongoConn.Create(r)
}

func (r *Rooms) List(filter room.NotifFilter, pfilter context.PublicFilter) (int, int, error) {
	q := filter.GetFilters()
	totalCount, err := config.MongoConn.Count(Collection, q)
	if err != nil {
		log.Error("can not count music list error : ", err.Error())
	}

	totalPages := totalCount / pfilter.Limit
	if totalPages*pfilter.Limit < totalCount {
		totalPages++
	}

	return totalCount, totalPages, config.MongoConn.Find(Collection_Room, q, r, pfilter.Limit, pfilter.Page, pfilter.Sort)
}

func (r *Room) LoadByID(id string) error {
	return config.MongoConn.FindOne(Collection_Room, bson.M{"id": id}, r)
}
