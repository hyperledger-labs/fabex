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

package main

import (
	"encoding/base64"

	"github.com/hyperledger-labs/fabex/config"
	"go.uber.org/zap"

	"github.com/hyperledger-labs/fabex/log"
	"github.com/hyperledger-labs/fabex/proto"

	fabcli "github.com/hyperledger-labs/fabex/client"
	"github.com/hyperledger-labs/fabex/helpers"
)

var (
	client *fabcli.FabexClient
	addr   = "0.0.0.0"
	port   = "6006"
)

func main() {
	l, err := log.GetLogger(&config.BootConfig{LogLevel: "info"})
	if err != nil {
		panic(err)
	}

	client, err = fabcli.New(addr, port)
	if err != nil {
		l.Panic(err.Error())
	}

	/*
	   Use this commented lines for your experiments!
	*/

	// get txs from blocks with block number range
	//txs, err := client.GetRange(1, 15)
	//if err != nil {
	//	l.Panic(err.Error())
	//}

	// get tx with tx ID
	//txs, err := client.Get(&pb.Entry{Txid:"3a3e933a3d9953b0b10e6573254b6d3cf2347d72058c0347a55054babdd8e1a1"})
	//if err != nil {
	//	l.Panic(err.Error())
	//}

	// get txs from specific block
	txs, err := client.Get(&proto.Entry{Channelid: "ch1", Blocknum: 1})
	if err != nil {
		l.Panic(err.Error())
	}

	// get entry with key
	//txs, err := client.Get(&proto.Entry{Channelid: "ch1", Payload: []byte("Policies")})
	//if err != nil {
	//	l.Panic(err.Error())
	//}

	// get all
	//txs, err := client.Get(&proto.Entry{Channelid: "ch1"})
	//if err != nil {
	//	l.Panic(err.Error())
	//}

	blocks, err := helpers.PackTxsToBlocks(txs)
	if err != nil {
		l.Panic(err.Error())
	}

	for _, block := range blocks {
		l.Info("found block", zap.Uint64("blocknum", block.Blocknum), zap.String("channel", block.ChannelId))

		for _, tx := range block.Txs {
			l.Info("tx", zap.String("ID", tx.Txid), zap.Int32("validation code", tx.ValidationCode))
			l.Info("Write set:")
			for _, write := range tx.KV {
				// decode base64-encoded value, coz bytes array can't be json serialized
				value, err := base64.StdEncoding.DecodeString(write.Value)
				if err != nil {
					l.Panic(err.Error())
				}
				l.Info("write set item", zap.String("key", write.Key), zap.String("value", string(value)))
			}
		}
	}
}
