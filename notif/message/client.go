package message

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageFilter struct {
	Room       string   `query:"room"`
	SearchKeys []string ``
	Search     string   `query:"search"`
}

func (filter MessageFilter) GetFilters() bson.M {
	q := bson.M{}

	if filter.Search != "" {
		if len(filter.SearchKeys) == 0 {
			filter.SearchKeys = []string{"title"}

		}
		orQ := []bson.M{}
		for _, key := range filter.SearchKeys {
			regex := bson.M{
				"$regex":   filter.Search,
				"$options": "i",
			}
			orQ = append(orQ, bson.M{key: regex})
		}
		if len(orQ) > 0 {
			q["$or"] = orQ
		}
	}

	if room, err := primitive.ObjectIDFromHex(filter.Room); err != nil {
		q["room_id"] = room
	}

	return q
}
