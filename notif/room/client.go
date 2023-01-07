package room

import "go.mongodb.org/mongo-driver/bson"

type NotifFilter struct {
	SearchKeys []string ``
	Search     string   `query:"search"`
}

func (filter NotifFilter) GetFilters() bson.M {
	q := bson.M{}

	if filter.Search != "" {
		if len(filter.SearchKeys) == 0 {
			filter.SearchKeys = []string{"keywords", "name_en", "name_fa"}

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

	return q
}
