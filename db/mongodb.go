package db

import (
	"context"
	pb "github.com/vadiminshakov/fabex/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
	"fmt"
)

type DBmongo struct {
	Host     string
	Port     int
	User     string
	Password string
	DBname   string
	Instance *mongo.Client
}

func CreateDBConfMongo(host string, port int, user, password, dbname string) *DBmongo {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Mongodb client creation failed: %s", err)
	}
	return &DBmongo{host, port, user, password, dbname, client}
}
func (db *DBmongo) Init() error {
	return nil
}

func (db *DBmongo) Connect() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := db.Instance.Connect(ctx)
	if err != nil {
		log.Println("Mongodb connection failed: %s", err)
		return err
	}
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = db.Instance.Ping(ctx, readpref.Primary())

	log.Println("Connected to MongoDB successfully")
	return nil
}

func (db *DBmongo) Insert(txid, blockhash string, blocknum uint64, payload []byte) error {
	collection := db.Instance.Database("blocks").Collection("txs")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"Txid": txid, "Blockhash": blockhash, "Blocknum": blocknum, "Payload": payload})
	log.Println(res.InsertedID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBmongo) QueryBlockByHash(hash string) (*QueryResult, error) {
	collection := db.Instance.Database("blocks").Collection("txs")
	var result *QueryResult
	filter := bson.M{"Blockhash": hash}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	// Do something with result...
	return &QueryResult{result.Txid, result.Blockhash, result.Blocknum, result.Payload}, nil
}

func (db *DBmongo) GetByTxId(filter *pb.RequestFilter) ([]*QueryResult, error) {
	collection := db.Instance.Database("blocks").Collection("txs")
	filterOpts := bson.M{"Txid": filter.Txid}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cur, err := collection.Find(ctx, filterOpts)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var results []*QueryResult
	for cur.Next(ctx) {
		var result *QueryResult
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
fmt.Println(results)
	return results, nil
}

func (db *DBmongo) QueryAll() ([]QueryResult, error) {
	arr := []QueryResult{}

	collection := db.Instance.Database("blocks").Collection("txs")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var (
			txid, blockhash string
			blocknum        int64
			payload         []byte
		)
		var result map[string]interface{}
		err := cur.Decode(&result)

		if err != nil {
			log.Fatal(err)
		}
		for k, v := range result {
			switch k {
			case "Txid":
				txid = v.(string)

			case "Blockhash":
				blockhash = v.(string)

			case "Blocknum":
				blocknum = v.(int64)

			case "Payload":
				bsonBinary := v.(primitive.Binary)
				payload = bsonBinary.Data
			}
		}
		var queryResultMongo = QueryResult{txid, blockhash, uint64(blocknum), payload}

		arr = append(arr, queryResultMongo)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return arr, nil
}
