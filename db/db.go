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

// Package db provides database interface for storing and retrieving blocks and transactions
package db

const NOT_FOUND_ERR = "not found"

// Storage db interface
type Storage interface {
	Connect() error
	Insert(tx Tx) error
	QueryBlockByHash(hash string) ([]Tx, error)
	GetByTxId(txid string) ([]Tx, error)
	GetByBlocknum(blocknum uint64) ([]Tx, error)
	GetBlockInfoByPayload(payload string) ([]Tx, error)
	QueryAll() ([]Tx, error)
	GetLastEntry() (Tx, error)
}

// Tx stores info about block and tx payload
type Tx struct {
	ChannelId      string `json:"channelid" bson:"ChannelId"`
	Txid           string `json:"txid" bson:"Txid"`
	Hash           string `json:"hash" bson:"Hash"`
	PreviousHash   string `json:"previoushash" bson:"PreviousHash"`
	Blocknum       uint64 `json:"blocknum" bson:"Blocknum"`
	Payload        []byte `json:"payload" bson:"Payload"`
	ValidationCode int32  `json:"validationcode" bson:"ValidationCode"`
	Time           int64  `json:"time" bson:"Time"`
}

// RW stores key and value of chaincode payload
type RW struct {
	Key   string
	Value string
}
