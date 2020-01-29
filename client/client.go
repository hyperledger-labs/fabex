package main

import (
	"fmt"
	"github.com/vadiminshakov/fabex/client/fabexclient"
	pb "github.com/vadiminshakov/fabex/proto"
	"log"
)

var (
	client *fabexclient.FabexClient
	addr   = "0.0.0.0"
	port   = "6000"
)

func init() {
	var err error
	client, err = fabexclient.New(addr, port)
	if err != nil {
		panic(err)
	}
}

func main() {
	//Explore(1, 15)
	//txs, err := client.GetByTxId(&pb.RequestFilter{Txid:"3a3e933a3d9953b0b10e6573254b6d3cf2347d72058c0347a55054babdd8e1a1"})
	//txs, err := client.GetByBlocknum(&pb.RequestFilter{Blocknum: 2})
	txs, err := client.GetBlockInfoByPayload(&pb.RequestFilter{Payload: "1440-"})
	if err != nil {
		log.Fatal(err)
	}

	blocks, err := client.PackTxsToBlocks(txs)
	fmt.Printf("%#v", blocks)
}
