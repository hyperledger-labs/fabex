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

type DbManager interface {
	Connect() error
	Init() error
	Insert(Tx) error
	QueryBlockByHash(hash string) ([]Tx, error)
	GetByTxId(string) ([]Tx, error)
	GetByBlocknum(uint64) ([]Tx, error)
	GetBlockInfoByPayload(string) ([]Tx, error)
	QueryAll() ([]Tx, error)
}

type Tx struct {
	ChannelId      string `json:"channelid" bson:"ChannelId"`
	Txid           string `json:"txid" bson:"Txid"`
	Hash           string `json:"hash" bson:"Hash"`
	PreviousHash   string `json:"previoushash" bson:"PreviousHash"`
	Blocknum       uint64 `json:"blocknum" bson:"Blocknum"`
	Payload        string `json:"payload" bson:"Payload"`
	ValidationCode int32  `json:"validationcode" bson:"ValidationCode"`
	Time           int64  `json:"time" bson:"Time"`
}

type Entry struct {
	ChannelId      string
	Txid           string
	Hash           string
	PreviousHash   string
	Blocknum       uint64
	Payload        string
	Time           int64
}
