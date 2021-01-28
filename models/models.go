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

package models

import (
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/ledgerclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

type Fabex struct {
	Db             db.Storage
	ChannelClient  *channel.Client
	LedgerClient   *ledgerclient.CustomLedgerClient
	ChannelContext contextApi.ChannelProvider
}

type Mongo struct {
	Host       string
	Port       int
	Dbuser     string
	Dbsecret   string
	Dbname     string
	Collection string
}

type Cassandra struct {
	Host         string
	Dbuser       string
	Dbsecret     string
	Keyspace     string
	Columnfamily string
}

type GRPCServer struct {
	Host string
	Port string
}

type Fabric struct {
	User              string
	Secret            string
	Org               string
	Channel           string
	ConnectionProfile string
}

type UI struct {
	Port string
}

type Config struct {
	Mongo
	Cassandra
	Fabric
	GRPCServer
	UI
}

type WriteKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Block struct {
	ChannelId    string `json:"channelid"`
	BlockHash    string `json:"blockhash"`
	PreviousHash string `json:"previoushash"`
	Blocknum     uint64 `json:"blocknum"`
	Txs          []Tx   `json:"txs"`
}

type Tx struct {
	Txid           string `json:"txid"`
	KV             []WriteKV
	ValidationCode int32 `json:"validationcode"`
	Time           int64 `json:"time" bson:"Time"`
}
