package db

import (
	"context"
	"fmt"
	pb "github.com/vadiminshakov/fabex/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
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

func (db *DBmongo) QueryBlockByHash(hash string) (Tx, error) {
	collection := db.Instance.Database("blocks").Collection("txs")
	var result Tx
	filter := bson.M{"Blockhash": hash}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return Tx{}, err
	}
	// Do something with result...
	return Tx{result.Txid, result.Hash, result.Blocknum, result.Payload}, nil
}

func (db *DBmongo) GetByTxId(filter *pb.RequestFilter) ([]Tx, error) {
	collection := db.Instance.Database("blocks").Collection("txs")
	filterOpts := bson.M{"Txid": filter.Txid}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cur, err := collection.Find(ctx, filterOpts)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var results []Tx
	for cur.Next(ctx) {
		var result Tx
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

func (db *DBmongo) GetByBlocknum(filter *pb.RequestFilter) ([]Tx, error) {
	collection := db.Instance.Database("blocks").Collection("txs")

	filterOpts := bson.M{"Blocknum": filter.Blocknum}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cur, err := collection.Find(ctx, filterOpts)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer cur.Close(ctx)

	var results []Tx
	for cur.Next(ctx) {
		var result Tx
		err = cur.Decode(&result)
		if err != nil {
			fmt.Println("ERR: ", err)
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *DBmongo) QueryAll() ([]Tx, error) {
	arr := []Tx{}

	collection := db.Instance.Database("blocks").Collection("txs")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result Tx
		err := cur.Decode(&result)

		if err != nil {
			log.Fatal(err)
		}

		arr = append(arr, result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return arr, nil
}
