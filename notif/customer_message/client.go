package customer_message

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerMessageFilter struct {
	Message    string   `query:"message"`
	Customer   string   `query:"customer"`
	SearchKeys []string ``
	Search     string   `query:"search"`
}

func (filter CustomerMessageFilter) GetFilters() bson.M {
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
	if message, err := primitive.ObjectIDFromHex(filter.Message); err == nil {
		q["message"] = message
	}

	if customer, err := primitive.ObjectIDFromHex(filter.Customer); err == nil {
		q["customer"] = customer
	}

	return q
}
