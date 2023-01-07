package mongo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultConfig = Config{
		DBName: "notification",
		URI:    "mongodb://localhost:27017",
	}
)

type Config struct {
	DBName string
	URI    string
}
type Conn struct {
	DB *mongo.Database
}

func InitMongoConnection(config Config) (*Conn, error) {
	opt := options.Client()
	opt.ApplyURI(config.URI)

	client, err := mongo.NewClient(opt)
	if err != nil {
		return nil, err
	}

	db := client.Database(config.DBName)
	if err := client.Connect(context.Background()); err != nil {
		return nil, errors.New("Connecting to MongoDB, " + err.Error())
	}
	conn := Conn{
		DB: db,
	}

	_, err = conn.DB.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return nil, errors.New("Listing Mongo Collection names, " + err.Error())
	}

	return &conn, nil
}

// Create a document in DBContext
func (v Conn) Create(model interface{}) error {
	col := v.DB.Collection(colName(model))
	setID(model)
	_, err := col.InsertOne(context.Background(), &model)
	return err
}

// Collection return mgo collection from model
func (v Conn) Collection(model interface{}) *mongo.Collection {
	return v.DB.Collection(colName(model))
}

// Collection return mgo collection from model
func (v Conn) CollectionString(model string) *mongo.Collection {
	return v.DB.Collection(model)
}

// Update a Document
func (v Conn) Update(model interface{}) error {
	collection := v.DB.Collection(colName(model))
	ctx := context.Background()
	id, err := getID(model)
	if err != nil {
		return err
	}
	query := bson.M{"_id": id}
	fieldsUpdate := parseBson(model)
	if _, err := collection.UpdateOne(ctx, query, bson.M{"$set": fieldsUpdate}); err != nil {
		return err
	}
	return v.Get(collection.Name(), id, model)
}

func (v Conn) UpdateMany(collection string, filter bson.M, update bson.M) error {
	ctx := context.Background()
	if _, err := v.DB.Collection(collection).UpdateMany(ctx, filter, update); err != nil {
		return err
	}
	return nil
}

func (v Conn) UpdateOne(collection string, filter bson.M, update bson.M) error {
	ctx := context.Background()
	if _, err := v.DB.Collection(collection).UpdateOne(ctx, filter, update); err != nil {
		return err
	}
	return nil
}

// Find start generating query
// order of options are:  limit, page, sort
func (v Conn) Find(collection string, q bson.M, result interface{}, params ...interface{}) error {
	col := v.DB.Collection(collection)
	sort := ""
	var page int64
	var limit int64
	opt := new(options.FindOptions)
	if len(params) > 0 {
		if l, ok := params[0].(int); ok {
			limit = int64(l)
		} else if lstr, ok := params[0].(string); ok {
			l, _ = strconv.Atoi(lstr)
			limit = int64(l)
		}
		if limit > 0 {
			opt.Limit = &limit
		}
	}
	if len(params) > 1 {
		if l, ok := params[1].(int); ok {
			page = int64(l)
		} else if lstr, ok := params[1].(string); ok {
			l, _ = strconv.Atoi(lstr)
			page = int64(l)
		}
		if page > 0 {
			skip := (page - 1) * limit
			opt.Skip = &skip
		}
	}
	if len(params) > 2 {
		sort = params[2].(string)
		// @TODO some ducument dont have created_at
		// if sort == "" {
		// 	sort = "-created_at"
		// }
		if sort != "" {
			if strings.HasPrefix(sort, "-") {
				sort = strings.ReplaceAll(sort, "-", "")
				opt.SetSort(bson.D{{Key: sort, Value: -1}})
			} else {
				opt.SetSort(bson.D{bson.E{Key: sort, Value: 1}})
			}
		}

	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := col.Find(ctx, q, opt)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(context.Background(), result)
	if err != nil {
		return err
	}
	return nil
}

func (v Conn) FindOne(collection string, q bson.M, result interface{}, params ...interface{}) error {
	col := v.DB.Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return col.FindOne(ctx, q).Decode(result)
}

func (v Conn) Count(collection string, filter bson.M) (int, error) {
	ctx := context.Background()
	c, err := v.DB.Collection(collection).CountDocuments(ctx, filter)
	return int(c), err
}

func (v Conn) RemoveOne(collection string, filter bson.M) error {
	ctx := context.Background()
	_, err := v.DB.Collection(collection).DeleteOne(ctx, filter)
	return err
}

func (v Conn) RemoveMany(collection string, filter bson.M) error {
	ctx := context.Background()
	_, err := v.DB.Collection(collection).DeleteMany(ctx, filter)
	return err
}

func (v Conn) Get(collection string, id interface{}, resultObj interface{}) error {
	if id == nil {
		return errors.New("invalid ID")
	}
	ctx := context.Background()
	var err error
	var objID primitive.ObjectID
	if val, ok := id.(string); ok {
		objID, err = primitive.ObjectIDFromHex(val)
		if err != nil {
			return err
		}
	} else {
		objID = id.(primitive.ObjectID)
	}
	if objID.IsZero() {
		return errors.New("invalid ID")
	}
	singleResult := v.DB.Collection(collection).FindOne(ctx, bson.M{"_id": objID})
	if singleResult.Err() != nil {
		return singleResult.Err()
	}
	return singleResult.Decode(resultObj)
}

func setID(model interface{}) {
	m := structs.Map(model)
	var keyID string
	if _, ok := m["Id"]; ok {
		m["Id"] = primitive.NewObjectID()
		keyID = "Id"
	}
	if _, ok := m["ID"]; ok {
		m["ID"] = primitive.NewObjectID()
		keyID = "ID"
	}
	s := structs.New(model)
	field := s.Field(keyID)
	field.Set(m[keyID])

}

func getID(model interface{}) (primitive.ObjectID, error) {
	m := structs.Map(model)
	var (
		idInterface interface{}
		id          primitive.ObjectID
		ok          bool
	)
	if val, ok := m["Id"]; ok {
		idInterface = val
	}
	if val, ok := m["ID"]; ok {
		idInterface = val
	}
	id, ok = idInterface.(primitive.ObjectID)
	if !ok {
		return id, errors.New("model id invalid")
	}
	if id.IsZero() {
		return id, errors.New("model id invalid")
	}
	return id, nil
}

func parseBson(model interface{}) bson.M {
	b, _ := bson.Marshal(model)
	var body bson.M
	bson.Unmarshal(b, &body)
	return body
}

type coller interface {
	CollectionName() string
}

// type indexer interface {
// 	Meta() []mongo.IndexModel
// }

func colName(model interface{}) string {
	if c, ok := model.(coller); ok {
		return c.CollectionName()
	}
	tmp := fmt.Sprintf("%T", model)
	tmp = strings.Replace(tmp, "*", "", -1)
	tmp = strings.Replace(tmp, "]", "", -1)
	tmp = strings.Replace(tmp, "[", "", -1)
	ts := strings.Split(tmp, ".")
	if len(ts) < 2 {
		return toSnake(tmp)
	}
	return toSnake(ts[1])
}

// func loadIndex(model interface{}) []mongo.IndexModel {
// 	if c, ok := model.(indexer); ok {
// 		return c.Meta()
// 	}
// 	return []mongo.IndexModel{}
// }

func toSnake(s string) string {
	var (
		res  string
		last int
	)
	ls := []rune(s)
	for i, char := range ls {
		if (i == 0 || !unicode.IsUpper(char)) && i+1 != len(s) {
			continue
		}
		if i+1 != len(s) {
			res += strings.ToLower(s[last:i]) + "_"
		} else {
			res += strings.ToLower(s[last : i+1])
		}
		last = i
	}
	return res
}
