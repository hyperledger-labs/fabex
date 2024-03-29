/*
   Copyright 2019 Vadim Inshakov

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBmongo struct {
	Host       string
	Port       int
	User       string
	Password   string
	DBname     string
	Collection string
	Instance   *mongo.Client
}

const ERR_NO_DOCUMENTS = "mongo: no documents in result"

func CreateDBConfMongo(host string, port int, user, password, dbname, collection string) *DBmongo {
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d", user, password, host, port)))
	if err != nil {
		log.Fatalf("Mongodb client creation failed: %s", err)
	}
	return &DBmongo{host, port, user, password, dbname, collection, client}
}

func (db *DBmongo) Connect() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) //nolint:govet
	err := db.Instance.Connect(ctx)
	if err != nil {
		return err
	}
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second) //nolint:govet
	err = db.Instance.Ping(ctx, readpref.Primary())
	return err
}

func (db *DBmongo) Init(_ string) error {
	return nil
}

func (db *DBmongo) Insert(ch string, tx Tx) error {
	collection := db.Instance.Database(db.DBname).Collection(fmt.Sprintf("%s_%s", db.Collection, ch))
	ctx := context.Background()

	_, err := collection.InsertOne(ctx, bson.M{"ChannelId": tx.ChannelId, "Txid": tx.Txid, "Hash": tx.Hash, "PreviousHash": tx.PreviousHash, "Blocknum": tx.Blocknum, "Payload": string(tx.Payload), "ValidationCode": tx.ValidationCode, "Time": tx.Time})
	if err != nil {
		return err
	}

	return nil
}

func (db *DBmongo) getByFilter(ch string, filterValue interface{}) ([]Tx, error) {
	collection := db.Instance.Database(db.DBname).Collection(fmt.Sprintf("%s_%s", db.Collection, ch))
	filter := filterValue
	ctx := context.Background()
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var results []Tx
	for cur.Next(ctx) {
		var result Tx
		err = cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *DBmongo) QueryBlockByHash(ch string, hash string) ([]Tx, error) {
	return db.getByFilter(ch, bson.M{"Hash": hash})
}

func (db *DBmongo) GetByTxId(ch string, txID string) ([]Tx, error) {
	return db.getByFilter(ch, bson.M{"Txid": txID})
}

func (db *DBmongo) GetByBlocknum(ch string, blocknum uint64) ([]Tx, error) {
	return db.getByFilter(ch, bson.M{"Blocknum": blocknum})
}

func (db *DBmongo) GetBlockInfoByPayload(ch string, payload string) ([]Tx, error) {
	return db.getByFilter(ch, bson.M{"Payload": primitive.Regex{Pattern: payload, Options: "i"}})
}

func (db *DBmongo) QueryAll(ch string) ([]Tx, error) {
	return db.getByFilter(ch, bson.D{})
}

func (db *DBmongo) GetLastEntry(ch string) (Tx, error) {
	collection := db.Instance.Database(db.DBname).Collection(fmt.Sprintf("%s_%s", db.Collection, ch))

	ctx := context.Background()
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})

	var tx Tx
	err := collection.FindOne(ctx, bson.D{}, opts).Decode(&tx)
	if err != nil && err.Error() != ERR_NO_DOCUMENTS {
		return tx, err
	} else if err != nil && err.Error() == ERR_NO_DOCUMENTS {
		return tx, errors.New(NOT_FOUND_ERR)
	}

	return tx, nil
}
