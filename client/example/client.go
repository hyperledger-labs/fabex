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
	fabcli "github.com/hyperledger-labs/fabex/client"
	"github.com/hyperledger-labs/fabex/helpers"
	"log"
)

var (
	client *fabcli.FabexClient
	addr   = "0.0.0.0"
	port   = "6000"
)

func main() {
	var err error
	client, err = fabcli.New(addr, port)
	if err != nil {
		panic(err)
	}

	/*
	   Use this commented lines for your experiments!
	*/

	// get txs from blocks with block number range
	//txs, err := client.GetRange(1, 15)

	// get tx with tx ID
	//txs, err := client.Get(&pb.Entry{Txid:"3a3e933a3d9953b0b10e6573254b6d3cf2347d72058c0347a55054babdd8e1a1"})

	// get txs from specific block
	//txs, err := client.Get(&pb.Entry{Blocknum: 5})

	// get entry with composite key
	//key, err := helpers.CreateCompositeKey("RAIL", []string{"1"})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//txs, err := client.Get(&pb.Entry{Payload: key})

	// get all
	txs, err := client.Get(nil)
	if err != nil {
		log.Fatal(err)
	}

	blocks, err := helpers.PackTxsToBlocks(txs)
	if err != nil {
		log.Fatal(err)
	}

	for _, block := range blocks {
		log.Printf("Blocknum: %d\nChannel:%s", block.Blocknum, block.ChannelId)
		for _, tx := range block.Txs {
			log.Printf("Tx ID: %s\nValidation code: %d", tx.Txid, tx.ValidationCode)
			log.Println("Write set:")
			for _, write := range tx.KV {
				// decode base64-encoded value, coz bytes array can't be json serialized
				value, err := base64.StdEncoding.DecodeString(write.Value)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Key: %s, Value: %s", write.Key, string(value))
			}
		}
	}
}
