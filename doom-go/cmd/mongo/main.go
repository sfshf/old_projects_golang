package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model1 struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	A  string             `bson:"a,omitempty" json:"a,omitempty"`
	B  string             `bson:"b,omitempty" json:"b,omitempty"`
	C  string             `bson:"c,omitempty" json:"c,omitempty"`
	D  string             `bson:"d,omitempty" json:"d,omitempty"`
}

type Model2 struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	A  string             `bson:"a,omitempty" json:"a,omitempty"`
	B  string             `bson:"b,omitempty" json:"b,omitempty"`
	D  int                `bson:"d,omitempty" json:"d,omitempty"`
	E  int                `bson:"e,omitempty" json:"e,omitempty"`
}

func queryAllByModel1(ctx context.Context, coll *mongo.Collection) error {
	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	for cursor.Next(ctx) {
		var m Model1
		if err := cursor.Decode(&m); err != nil {
			return err
		}
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		log.Println(string(data))
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	return nil
}

func queryAllByModel2(ctx context.Context, coll *mongo.Collection) error {
	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	for cursor.Next(ctx) {
		var m Model2
		if err := cursor.Decode(&m); err != nil {
			return err
		}
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		log.Println(string(data))
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	start := time.Now()
	log.Println("start time:", start.Format("2006-01-02 15:04:05"))
	ctx := context.Background()
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Println(err)
		return
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
	}
	// init mongo client
	mgoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
	if err != nil {
		log.Println(err)
		return
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Println(err)
		return
	}
	log.Println("mongodb connect successfully")
	// init collection
	coll := mgoCli.Database(dbName).Collection("mongo_test")
	defer func() {
		if err := coll.Drop(ctx); err != nil {
			log.Println(err)
			return
		}
	}()
	// 5 records
	if _, err := coll.InsertMany(
		ctx,
		[]interface{}{
			Model1{A: "a1", B: "b1", C: "c1", D: "d1"},
			Model1{A: "a2", B: "b2", C: "c2", D: "d2"},
			Model1{A: "a3", B: "b3", C: "c3", D: "d3"},
			Model1{A: "a4", B: "b4", C: "c4", D: "d4"},
			Model1{A: "a5", B: "b5", C: "c5", D: "d5"},
		}); err != nil {
		log.Println(err)
		return
	}
	log.Println("After insert 5 records:")
	if err := queryAllByModel1(ctx, coll); err != nil {
		log.Println(err)
		return
	}
	// update collection schema
	// 删掉 c 字段
	if _, err := coll.UpdateMany(ctx, bson.D{}, bson.D{{Key: "$unset", Value: bson.D{{Key: "c", Value: ""}}}}); err != nil {
		log.Println(err)
		return
	}
	log.Println("After unset c field:")
	if err := queryAllByModel1(ctx, coll); err != nil {
		log.Println(err)
		return
	}
	// 修改d字段类型，为number
	cursor, err := coll.Find(ctx, bson.D{{Key: "d", Value: bson.D{{Key: "$exists", Value: true}, {Key: "$type", Value: "string"}}}})
	if err != nil {
		log.Println(err)
		return
	}
	for cursor.Next(ctx) {
		var m1 Model1
		if err := cursor.Decode(&m1); err != nil {
			log.Println(err)
			return
		}
		m2 := Model2{
			ID: m1.ID,
			A:  m1.A,
			B:  m1.B,
		}
		if m1.D != "" {
			m2.D = 2
		} else {
			m2.D = 1
		}
		if _, err := coll.UpdateOne(ctx, bson.D{{Key: "_id", Value: m1.ID}}, bson.D{{Key: "$set", Value: m2}}); err != nil {
			log.Println(err)
			return
		}
	}
	if err := cursor.Err(); err != nil {
		log.Println(err)
		return
	}
	log.Println("After modify d field:")
	if err := queryAllByModel2(ctx, coll); err != nil {
		log.Println(err)
		return
	}
	// 添加 e字段， 类型为 number
	if _, err := coll.UpdateMany(ctx, bson.D{}, bson.D{{Key: "$set", Value: bson.D{{Key: "e", Value: 0}}}}); err != nil {
		log.Println(err)
		return
	}
	log.Println("After set e field:")
	if err := queryAllByModel2(ctx, coll); err != nil {
		log.Println(err)
		return
	}
	// 5 records
	if _, err := coll.InsertMany(
		ctx,
		[]interface{}{
			Model2{A: "a1", B: "b1", D: 0xd1, E: 0xe1},
			Model2{A: "a2", B: "b2", D: 0xd2, E: 0xe1},
			Model2{A: "a3", B: "b3", D: 0xd3, E: 0xe1},
			Model2{A: "a4", B: "b4", D: 0xd4, E: 0xe1},
			Model2{A: "a5", B: "b5", D: 0xd5, E: 0xe1},
		}); err != nil {
		log.Println(err)
		return
	}
	log.Println("After insert another 5 records:")
	if err := queryAllByModel2(ctx, coll); err != nil {
		log.Println(err)
		return
	}
	end := time.Now()
	log.Println("end time:", end.Format("2006-01-02 15:04:05"))
	log.Println("duration:", end.Sub(start).String())
}
