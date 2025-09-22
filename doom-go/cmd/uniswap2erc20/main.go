package main

import (
	"context"
	"log"
	"net/url"
	"time"

	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UniswapV2FactoryAddress = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
)

func uniswap2erc20(ctx context.Context, uniswapCollection, erc20Collection *mongo.Collection) error {
	tokenChannel := make(chan UniswapTokens, 50)
	for i := 0; i < 50; i++ {
		go func() {
			for m := range tokenChannel {
				filter := bson.D{{Key: "key", Value: m.Key}}
				if m.Value.Type == UniswapTokenTypeV2 {
					filter = append(filter, bson.E{Key: "value.symbol", Value: "UNI-V2"})
				} else if m.Value.Type == UniswapTokenTypeV3 {
					filter = append(filter, bson.E{Key: "value.symbol", Value: "UNI-V3"})
				}
				cnt, err := erc20Collection.CountDocuments(ctx, filter)
				if err != nil {
					log.Println(err)
					return
				}
				if cnt == 1 {
					continue
				}
				erc20Token := Erc20Tokens{
					Key: m.Key,
					Value: Erc20Tokens_Value{
						Type:     TokenTypeERC20,
						Decimals: 18,
					},
				}
				if m.Value.Type == UniswapTokenTypeV2 {
					erc20Token.Value.Name = "Uniswap V2"
					erc20Token.Value.Symbol = "UNI-V2"
				} else if m.Value.Type == UniswapTokenTypeV3 {
					erc20Token.Value.Name = "Uniswap V3"
					erc20Token.Value.Symbol = "UNI-V3"
				}
				if _, err := erc20Collection.ReplaceOne(ctx, bson.D{{Key: "key", Value: erc20Token.Key}}, erc20Token, options.Replace().SetUpsert(true)); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	}
	cursor, err := uniswapCollection.Find(ctx, bson.D{{Key: "value.type", Value: bson.D{{Key: "$exists", Value: true}}}}, options.Find().SetBatchSize(2000))
	if err != nil {
		log.Println(err)
		return err
	}
	for cursor.Next(ctx) {
		var m UniswapTokens
		if err := cursor.Decode(&m); err != nil {
			log.Println(err)
			return err
		}
		tokenChannel <- m
	}
	time.Sleep(30 * time.Second)
	close(tokenChannel)
	return cursor.Err()
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	ctx := context.Background()
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Fatalln(err)
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
	}
	// init mongo client
	mgoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
	if err != nil {
		log.Fatalln(err)
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Fatalln(err)
	}
	log.Println("mongodb connect successfully")
	// init collection
	uniswapCollection := mgoCli.Database(dbName).Collection(CollectionName_UniswapTokens)
	erc20Collection := mgoCli.Database(dbName).Collection(CollectionName_ERC20Tokens)
	if err := uniswap2erc20(ctx, uniswapCollection, erc20Collection); err != nil {
		log.Fatalln(err)
	}
	log.Println("ok")
}
