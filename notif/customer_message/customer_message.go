package customer_message

import (
	"encoding/json"
	"notification/config"
	"notification/context"
	"notification/notif/message"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	CustomerMessage struct {
		ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
		Customer primitive.ObjectID `bson:"customer" json:"customer"`
		Message  primitive.ObjectID `bson:"message" json:"message"`
	}
	CustomerMessages []CustomerMessage
)

const Collection = "customer_message"

func (cm *CustomerMessage) Rest() echo.Map {
	resp := map[string]interface{}{}

	bytes, _ := json.Marshal(cm)
	json.Unmarshal(bytes, &resp)

	/* 	c := new(customer.Customer)
	   	c.LoadByID(cm.Customer)
	   	resp["customer"] = c.Rest() */

	msg := new(message.Message)
	msg.LoadByID(cm.Message)
	resp["message"] = msg

	return resp
}

func (c_ms *CustomerMessages) Rest() []echo.Map {
	resp := make([]echo.Map, 0)
	for _, r := range *c_ms {
		resp = append(resp, r.Rest())
	}

	return resp
}

func (c_m *CustomerMessage) Save() error {
	return config.MongoConn.Create(c_m)
}

func (c_ms *CustomerMessages) List(filter CustomerMessageFilter, pfilter context.PublicFilter) (int, int, error) {
	q := filter.GetFilters()
	totalCount, err := config.MongoConn.Count(Collection, q)
	if err != nil {
		log.Error("can not count message list error : ", err.Error())
	}

	totalPages := totalCount / pfilter.Limit
	if totalPages*pfilter.Limit < totalCount {
		totalPages++
	}

	return totalCount, totalPages, config.MongoConn.Find(Collection, q, c_ms, pfilter.Limit, pfilter.Page, pfilter.Sort)
}
